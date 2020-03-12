package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/transport"
	"github.com/cloudevents/sdk-go/pkg/transport/bindings"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
)

type Transport struct {
	bindings.BindingTransport

	// The encoding used to select the codec for outbound events.
	Encoding Encoding

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
	transport         http.RoundTripper // TODO: use this.
	server            *http.Server
	handlerRegistered bool
	middleware        []Middleware

	protocol *Protocol
}

func New(p *Protocol, opts ...Option) (*Transport, error) {
	t := &Transport{}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	t.Sender = p
	t.Requester = p
	t.Receiver = p
	t.Responder = p
	t.protocol = p

	if t.ShutdownTimeout == nil {
		timeout := DefaultShutdownTimeout
		t.ShutdownTimeout = &timeout
	}

	return t, nil
}

// Protocol returns a CloudEvents transport.Protocol.
// Deprecated: This is for legacy transition until client uses Sender/Receiver directly.
func (t *Transport) Transport() transport.Transport {
	return t
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
func (t *Transport) Send(ctx context.Context, e event.Event) error {
	switch t.Encoding {
	case Default, Binary:
		ctx = binding.WithForceBinary(ctx)
	case Structured:
		ctx = binding.WithForceStructured(ctx)
	}

	return t.BindingTransport.Send(ctx, e)
}

// Request implements Transport.Request
func (t *Transport) Request(ctx context.Context, e event.Event) (*event.Event, error) {
	switch t.Encoding {
	case Default, Binary:
		ctx = binding.WithForceBinary(ctx)
	case Structured:
		ctx = binding.WithForceStructured(ctx)
	}

	return t.BindingTransport.Request(ctx, e)
}

// HasTracePropagation implements Protocol.HasTracePropagation
func (t *Transport) HasTracePropagation() bool {
	return false
}

// StartReceiver implements Transport.StartReceiver
// NOTE: This is a blocking call.
func (t *Transport) StartReceiver(ctx context.Context) error {
	t.reMu.Lock()
	defer t.reMu.Unlock()

	if t.Handler == nil {
		t.Handler = http.NewServeMux()
	}

	if !t.handlerRegistered {
		// handler.Handle might panic if the user tries to use the same path as the sdk.
		t.Handler.Handle(t.GetPath(), t.protocol)
		t.handlerRegistered = true
	}

	addr, err := t.listen()
	if err != nil {
		return err
	}

	t.server = &http.Server{
		Addr: addr.String(),
		Handler: &ochttp.Handler{
			Propagation:    &tracecontext.HTTPFormat{},
			Handler:        attachMiddleware(t.Handler, t.middleware),
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
		if err := t.BindingTransport.StartReceiver(ctx); err != nil {
			errChan <- err
		}
	}()

	// wait for the server to return or ctx.Done().
	select {
	case <-ctx.Done():
		// Try a gracefully shutdown.
		timeout := *t.ShutdownTimeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err := t.server.Shutdown(ctx)
		<-errChan // Wait for server goroutine to exit
		return err
	case err := <-errChan:
		return err
	}
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

//
//func formatSpanName(r *http.Request) string {
//	return "cloudevents.http." + r.URL.Path
//}

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

//
//// attachMiddleware attaches the HTTP middleware to the specified handler.
//func attachMiddleware(h http.Handler, middleware []Middleware) http.Handler {
//	for _, m := range middleware {
//		h = m(h)
//	}
//	return h
//}
