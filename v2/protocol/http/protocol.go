package http

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/cloudevents/sdk-go/v2/types"

	"github.com/cloudevents/sdk-go/v2/protocol"

	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
)

const (
	// DefaultShutdownTimeout defines the default timeout given to the http.Server when calling Shutdown.
	DefaultShutdownTimeout = time.Minute * 1
)

// Protocol acts as both a http client and a http handler.
type Protocol struct {
	Target          *url.URL
	RequestTemplate *http.Request
	transformers    binding.TransformerFactories
	Client          *http.Client
	incoming        chan msgErr

	// To support Opener:

	// ShutdownTimeout defines the timeout given to the http.Server when calling Shutdown.
	// If nil, DefaultShutdownTimeout is used.
	ShutdownTimeout *time.Duration

	// Port is the port to bind the receiver to. Defaults to 8080.
	Port *int
	// Path is the path to bind the receiver to. Defaults to "/".
	Path string

	// Receive Mutex
	reMu sync.Mutex
	// Handler is the handler the http Server will use. Use this to reuse the
	// http server. If nil, the Protocol will create a one.
	Handler           *http.ServeMux
	listener          net.Listener
	roundTripper      http.RoundTripper
	server            *http.Server
	handlerRegistered bool
	middleware        []Middleware
}

func New(opts ...Option) (*Protocol, error) {
	p := &Protocol{
		transformers: make(binding.TransformerFactories, 0),
		incoming:     make(chan msgErr),
	}
	if err := p.applyOptions(opts...); err != nil {
		return nil, err
	}

	if p.Client == nil {
		p.Client = http.DefaultClient
	}

	if p.roundTripper != nil {
		p.Client.Transport = p.roundTripper
	}

	if p.ShutdownTimeout == nil {
		timeout := DefaultShutdownTimeout
		p.ShutdownTimeout = &timeout
	}

	return p, nil
}

func (p *Protocol) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(p); err != nil {
			return err
		}
	}
	return nil
}

// Send implements binding.Sender
func (p *Protocol) Send(ctx context.Context, m binding.Message) error {
	if ctx == nil {
		return fmt.Errorf("nil Context")
	} else if m == nil {
		return fmt.Errorf("nil Message")
	}

	_, err := p.Request(ctx, m)
	return err
}

// Request implements binding.Requester
func (p *Protocol) Request(ctx context.Context, m binding.Message) (binding.Message, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil Context")
	} else if m == nil {
		return nil, fmt.Errorf("nil Message")
	}

	var err error
	defer func() { _ = m.Finish(err) }()

	req := p.makeRequest(ctx)

	if p.Client == nil || req == nil || req.URL == nil {
		return nil, fmt.Errorf("not initialized: %#v", p)
	}

	if err = WriteRequest(ctx, m, req, p.transformers); err != nil {
		return nil, err
	}

	do := p.doOnce
	if backoff := cecontext.RetryFrom(ctx); backoff != nil {
		do = p.doWithRetry(backoff)
	}

	return do(req)
}

func (p *Protocol) doOnce(req *http.Request) (binding.Message, error) {
	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, protocol.NewReceipt(false, "%w", err)
	}

	var result protocol.Result
	if resp.StatusCode/100 == 2 {
		result = protocol.ResultACK
	} else {
		result = protocol.ResultNACK
	}

	return NewMessage(resp.Header, resp.Body), NewResult(resp.StatusCode, "%w", result)
}

func (p *Protocol) doWithRetry(backoff *types.Backoff) func(*http.Request) (binding.Message, error) {
	return func(req *http.Request) (binding.Message, error) {
		retries := 0
		var results []protocol.Result
		for {
			resp, err := p.Client.Do(req)

			// Fast track common case.
			if err == nil && resp.StatusCode/100 == 2 {
				result := NewResult(resp.StatusCode, "%w", protocol.ResultACK)

				if retries != 0 {
					result = NewRetriesResult(result, retries, results)
				}
				return NewMessage(resp.Header, resp.Body), result
			}

			// Slow case.
			retry := retries < backoff.Retry

			var result protocol.Result
			if err != nil {
				result = protocol.NewReceipt(false, "%w", err)
				if retry && !err.(*url.Error).Timeout() {
					// Do not retry if the error is not a timeout
					retry = false
				}
			} else {
				// No error, status code is not 2xx
				result = NewResult(resp.StatusCode, "%w", protocol.ResultNACK)

				// Potentially retry when:
				// - 413 Payload Too Large with Retry-After (NOT SUPPORTED)
				// - 425 Too Early
				// - 429 Too Many Requests
				// - 503 Service Unavailable (with or without Retry-After) (IGNORE Retry-After)
				// - 504 Gateway Timeout

				sc := resp.StatusCode
				if retry && sc != 425 && sc != 429 && sc != 503 && sc != 504 {
					// Permanent error
					retry = false
				}
			}

			if !retry {
				return NewMessage(resp.Header, resp.Body), NewRetriesResult(result, retries, results)
			}

			results = append(results, result)
			retries++

			// Linear backoff.
			time.Sleep(backoff.Delay)
		}
	}
}

func (p *Protocol) makeRequest(ctx context.Context) *http.Request {
	// TODO: support custom headers from context?
	req := &http.Request{
		Method: http.MethodPost,
		Header: make(http.Header),
		// TODO: HeaderFrom(ctx),
	}

	if p.RequestTemplate != nil {
		req.Method = p.RequestTemplate.Method
		req.URL = p.RequestTemplate.URL
		req.Close = p.RequestTemplate.Close
		req.Host = p.RequestTemplate.Host
		copyHeadersEnsure(p.RequestTemplate.Header, &req.Header)
	}

	if p.Target != nil {
		req.URL = p.Target
	}

	// Override the default request with target from context.
	if target := cecontext.TargetFrom(ctx); target != nil {
		req.URL = target
	}
	return req.WithContext(ctx)
}

// Ensure to is a non-nil map before copying
func copyHeadersEnsure(from http.Header, to *http.Header) {
	if len(from) > 0 {
		if *to == nil {
			*to = http.Header{}
		}
		copyHeaders(from, *to)
	}
}

func copyHeaders(from, to http.Header) {
	if from == nil || to == nil {
		return
	}
	for header, values := range from {
		for _, value := range values {
			to.Add(header, value)
		}
	}
}

// Receive the next incoming HTTP request as a CloudEvent.
// Returns non-nil error if the incoming HTTP request fails to parse as a CloudEvent
// Returns io.EOF if the receiver is closed.
func (p *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil Context")
	}

	msg, fn, err := p.Respond(ctx)
	// No-op the response when finish is invoked.
	if msg != nil {
		return binding.WithFinish(msg, func(err error) {
			if fn != nil {
				_ = fn(ctx, nil, nil)
			}
		}), err
	} else {
		return nil, err
	}
}

// Respond receives the next incoming HTTP request as a CloudEvent and waits
// for the response callback to invoked before continuing.
// Returns non-nil error if the incoming HTTP request fails to parse as a CloudEvent
// Returns io.EOF if the receiver is closed.
func (p *Protocol) Respond(ctx context.Context) (binding.Message, protocol.ResponseFn, error) {
	if ctx == nil {
		return nil, nil, fmt.Errorf("nil Context")
	}

	select {
	case in, ok := <-p.incoming:
		if !ok {
			return nil, nil, io.EOF
		}
		return in.msg, in.respFn, in.err
	case <-ctx.Done():
		return nil, nil, io.EOF
	}
}

type msgErr struct {
	msg    *Message
	respFn protocol.ResponseFn
	err    error
}

// ServeHTTP implements http.Handler.
// Blocks until ResponseFn is invoked.
func (p *Protocol) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m := NewMessageFromHttpRequest(req)
	if m == nil || m.ReadEncoding() == binding.EncodingUnknown {
		p.incoming <- msgErr{msg: nil, err: binding.ErrUnknownEncoding}
		return // if there was no message, return.
	}

	done := make(chan struct{})
	var finishErr error

	m.OnFinish = func(err error) error {
		finishErr = err
		return nil
	}

	var fn protocol.ResponseFn = func(ctx context.Context, resp binding.Message, er protocol.Result) error {
		// Unblock the ServeHTTP after the reply is written
		defer func() {
			done <- struct{}{}
		}()
		status := http.StatusOK
		if finishErr != nil {
			http.Error(rw, fmt.Sprintf("cannot forward CloudEvent: %v", finishErr), http.StatusInternalServerError)
		}
		if er != nil {
			var result *Result
			if protocol.ResultAs(er, &result) {
				if result.StatusCode > 100 && result.StatusCode < 600 {
					status = result.StatusCode
				}
			}
		}
		if resp != nil {
			err := WriteResponseWriter(ctx, resp, status, rw, p.transformers)
			return resp.Finish(err)
		}
		rw.WriteHeader(status)
		return nil
	}

	p.incoming <- msgErr{msg: m, respFn: fn} // Send to Request
	// Block until ResponseFn is invoked
	<-done
}
