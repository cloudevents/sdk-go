package httpb

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/bindings/http"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.uber.org/zap"
	"io"
	"net"
	nethttp "net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/cloudevents/sdk-go/pkg/event"
)

// Transport adheres to transport.Transport.
var _ transport.Transport = (*Transport)(nil)

const (
	// DefaultShutdownTimeout defines the default timeout given to the http.Server when calling Shutdown.
	DefaultShutdownTimeout = time.Minute * 1
)

// Transport acts as both a http client and a http handler.
type Transport struct {
	sender   binding.Sender
	receiver *http.Receiver // implements binding.Receiver

	consumer transport.Receiver

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
	// http server. If nil, the Transport will create a one.
	Handler           *nethttp.ServeMux
	listener          net.Listener
	server            *nethttp.Server
	handlerRegistered bool
}

// Option is the function signature required to be considered an http.Option.
type Option func(*Transport) error

func New(opts ...Option) (*Transport, error) {
	t := &Transport{}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	if t.sender == nil {
		client := nethttp.DefaultClient
		target, _ := url.Parse("http://localhost:8080")
		t.sender = http.NewSender(client, target)
	}

	return t, nil
}

func (t *Transport) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

// Send implements Transport.Send
func (t *Transport) Send(ctx context.Context, e event.Event) (context.Context, *event.Event, error) {
	msg := binding.EventMessage(e)

	if err := t.sender.Send(ctx, msg); err != nil {
		return ctx, nil, err
	}

	return ctx, nil, nil
}

// SetReceiver implements Transport.SetReceiver
func (t *Transport) SetReceiver(r transport.Receiver) {
	t.consumer = r
}

// SetConverter implements Transport.SetConverter
func (t *Transport) SetConverter(c transport.Converter) {
	// TODO: implement converter
	panic("not implemented")
}

// HasConverter implements Transport.HasConverter
func (t *Transport) HasConverter() bool {
	// TODO: implement converter
	return false
}

// StartReceiver implements Transport.StartReceiver
// NOTE: This is a blocking call.
func (t *Transport) StartReceiver(ctx context.Context) error {
	t.reMu.Lock()
	defer t.reMu.Unlock()

	logger := cecontext.LoggerFrom(ctx)

	if t.receiver == nil {
		t.receiver = http.NewReceiver()
	}

	if t.Handler == nil {
		t.Handler = nethttp.NewServeMux()
	}

	if !t.handlerRegistered {
		// handler.Handle might panic if the user tries to use the same path as the sdk.
		t.Handler.Handle(t.GetPath(), t.receiver)
		t.handlerRegistered = true
	}

	addr, err := t.listen()
	if err != nil {
		return err
	}

	t.server = &nethttp.Server{
		Addr: addr.String(),
		Handler: &ochttp.Handler{
			Propagation: &tracecontext.HTTPFormat{},
			// TODO: support middleware
			//Handler:        attachMiddleware(t.Handler, t.middleware),
			Handler:        t.Handler,
			FormatSpanName: formatSpanName,
		},
	}

	// Shutdown
	defer func() {
		_ = t.server.Close()
		t.server = nil
	}()

	errChan := make(chan error, 1)
	go func() {
		errChan <- t.server.Serve(t.listener)
	}()

	go func() {
		for {
			msg, err := t.receiver.Receive(ctx)
			if err == io.EOF {
				// shut it down.
				errChan <- err
				break
			} else if err != nil {
				logger.Errorw("failed to receive:", zap.Error(err))
				continue
			}
			e, _, err := binding.ToEvent(ctx, msg)

			// TODO: response is not supported in PoC.
			_, err = t.invokeReceiver(ctx, e)

			err = msg.Finish(err)
			if err != nil {
				logger.Errorw("failed to Finish message: ", zap.Error(err))
			}
		}
	}()

	// wait for the server to return or ctx.Done().
	select {
	case <-ctx.Done():
		// Try a gracefully shutdown.
		timeout := DefaultShutdownTimeout
		if t.ShutdownTimeout != nil {
			timeout = *t.ShutdownTimeout
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err := t.server.Shutdown(ctx)
		<-errChan // Wait for server goroutine to exit
		return err
	case err := <-errChan:
		return err
	}
}

// TODO: this is not what we want to use long term. For PoC.
// Response is an http transport response.
type Response struct {
	StatusCode int
	Message
}

// Message is an http transport message.
type Message struct {
	Header nethttp.Header
	Body   []byte
}

func (t *Transport) invokeReceiver(ctx context.Context, e event.Event) (*Response, error) {
	logger := cecontext.LoggerFrom(ctx)
	if t.consumer != nil {
		// Note: http does not use eventResp.Reason
		eventResp := event.EventResponse{}
		resp := Response{}

		err := t.consumer.Receive(ctx, e, &eventResp)
		if err != nil {
			logger.Warnw("got an error from receiver fn", zap.Error(err))
			resp.StatusCode = nethttp.StatusInternalServerError
			return &resp, err
		}

		// TODO: response is not supported in the PoC.
		//if eventResp.Event != nil {
		//	if t.loadCodec(ctx) {
		//		if m, err := t.codec.Encode(ctx, *eventResp.Event); err != nil {
		//			logger.Errorw("failed to encode response from receiver fn", zap.Error(err))
		//		} else if msg, ok := m.(*Message); ok {
		//			resp.Message = *msg
		//		}
		//	} else {
		//		logger.Error("failed to load codec")
		//		resp.StatusCode = http.StatusInternalServerError
		//		return &resp, err
		//	}
		//	// Look for a transport response context
		//	var trx *TransportResponseContext
		//	if ptrTrx, ok := eventResp.Context.(*TransportResponseContext); ok {
		//		// found a *TransportResponseContext, use it.
		//		trx = ptrTrx
		//	} else if realTrx, ok := eventResp.Context.(TransportResponseContext); ok {
		//		// found a TransportResponseContext, make it a pointer.
		//		trx = &realTrx
		//	}
		//	// If we found a TransportResponseContext, use it.
		//	if trx != nil && trx.Header != nil && len(trx.Header) > 0 {
		//		copyHeadersEnsure(trx.Header, &resp.Message.Header)
		//	}
		//}

		if eventResp.Status != 0 {
			resp.StatusCode = eventResp.Status
		} else {
			resp.StatusCode = nethttp.StatusAccepted // default is 202 - Accepted
		}
		return &resp, err
	}
	return nil, nil
}

// HasTracePropagation implements Transport.HasTracePropagation
func (t *Transport) HasTracePropagation() bool {
	return false
}

// GetPort returns the listening port.
// Returns -1 if there is a listening error.
// Note this will call net.Listen() if  the listener is not already started.
func (t *Transport) GetPort() int {
	// Ensure we have a listener and therefore a port.
	if _, err := t.listen(); err == nil || t.Port != nil {
		return *t.Port
	}
	return -1
}

func formatSpanName(r *nethttp.Request) string {
	return "cloudevents.http." + r.URL.Path
}

func (t *Transport) setPort(port int) {
	if t.Port == nil {
		t.Port = new(int)
	}
	*t.Port = port
}

// listen if not already listening, update t.Port
func (t *Transport) listen() (net.Addr, error) {
	if t.listener == nil {
		port := 8080
		if t.Port != nil {
			port = *t.Port
			if port < 0 || port > 65535 {
				return nil, fmt.Errorf("invalid port %d", port)
			}
		}
		var err error
		if t.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
			return nil, err
		}
	}
	addr := t.listener.Addr()
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		t.setPort(tcpAddr.Port)
	}
	return addr, nil
}

// GetPath returns the path the transport is hosted on. If the path is '/',
// the transport will handle requests on any URI. To discover the true path
// a request was received on, inspect the context from Receive(cxt, ...) with
// TransportContextFrom(ctx).
func (t *Transport) GetPath() string {
	path := strings.TrimSpace(t.Path)
	if len(path) > 0 {
		return path
	}
	return "/" // default
}
