package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cloudevents/sdk-go/pkg/transport"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
)

var _ transport.Engine = (*Engine)(nil)

type Engine struct {
	// The encoding used to select the codec for outbound events.
	//	Encoding Encoding

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
}

func (e *Engine) Inbound(ctx context.Context, inbound interface{}) error {
	handler, ok := inbound.(http.Handler)
	if !ok {
		return errors.New("protocol must implement http.Handler")
	}

	e.reMu.Lock()
	defer e.reMu.Unlock()

	if e.Handler == nil {
		e.Handler = http.NewServeMux()
	}

	if !e.handlerRegistered {
		// handler.Handle might panic if the user tries to use the same path as the sdk.
		e.Handler.Handle(e.GetPath(), handler)
		e.handlerRegistered = true
	}

	addr, err := e.listen()
	if err != nil {
		return err
	}

	e.server = &http.Server{
		Addr: addr.String(),
		Handler: &ochttp.Handler{
			Propagation:    &tracecontext.HTTPFormat{},
			Handler:        attachMiddleware(e.Handler, e.middleware),
			FormatSpanName: formatSpanName,
		},
	}

	// Shutdown
	defer func() {
		_ = e.server.Close()
		e.server = nil
	}()

	errChan := make(chan error, 1)
	go func() {
		errChan <- e.server.Serve(e.listener)
	}()

	//go func() {
	//	if err := e.BindingTransport.StartReceiver(ctx); err != nil {
	//		errChan <- err
	//	}
	//}()

	// wait for the server to return or ctx.Done().
	select {
	case <-ctx.Done():
		// Try a gracefully shutdown.
		timeout := *e.ShutdownTimeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err := e.server.Shutdown(ctx)
		<-errChan // Wait for server goroutine to exit
		return err
	case err := <-errChan:
		return err
	}
}

func (e *Engine) Outbound(ctx context.Context, outbound interface{}) error {
	panic("implement me")
}

func NewEngine(opts ...EngineOption) (*Engine, error) {
	t := &Engine{}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	if t.ShutdownTimeout == nil {
		timeout := DefaultShutdownTimeout
		t.ShutdownTimeout = &timeout
	}

	return t, nil
}

func (e *Engine) applyOptions(opts ...EngineOption) error {
	for _, fn := range opts {
		if err := fn(e); err != nil {
			return err
		}
	}
	return nil
}

// HasTracePropagation implements Protocol.HasTracePropagation
func (e *Engine) HasTracePropagation() bool { // TODO: clean this all up.
	return false
}

// GetPort returns the listening port.
// Returns -1 if there is a listening error.
// Note this will call net.Listen() if  the listener is not already started.
func (e *Engine) GetPort() int {
	// Ensure we have a listener and therefore a port.
	if _, err := e.listen(); err == nil || e.Port != nil {
		return *e.Port
	}
	return -1
}

func formatSpanName(r *http.Request) string {
	return "cloudevents.http." + r.URL.Path
}

func (e *Engine) setPort(port int) {
	if e.Port == nil {
		e.Port = new(int)
	}
	*e.Port = port
}

// listen if not already listening, update t.Port
func (e *Engine) listen() (net.Addr, error) {
	if e.listener == nil {
		port := 8080
		if e.Port != nil {
			port = *e.Port
			if port < 0 || port > 65535 {
				return nil, fmt.Errorf("invalid port %d", port)
			}
		}
		var err error
		if e.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
			return nil, err
		}
	}
	addr := e.listener.Addr()
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		e.setPort(tcpAddr.Port)
	}
	return addr, nil
}

// GetPath returns the path the transport is hosted on. If the path is '/',
// the transport will handle requests on any URI. To discover the true path
// a request was received on, inspect the context from Receive(cxt, ...) with
// TransportContextFrom(ctx).
func (e *Engine) GetPath() string {
	path := strings.TrimSpace(e.Path)
	if len(path) > 0 {
		return path
	}
	return "/" // default
}

// attachMiddleware attaches the HTTP middleware to the specified handler.
func attachMiddleware(h http.Handler, middleware []Middleware) http.Handler {
	for _, m := range middleware {
		h = m(h)
	}
	return h
}
