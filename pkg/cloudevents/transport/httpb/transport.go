package httpb

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/bindings/http"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
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
	binding.BindingTransport

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
	// http server. If nil, the Transport will create a one.
	Handler           *nethttp.ServeMux
	listener          net.Listener
	server            *nethttp.Server
	handlerRegistered bool
	middleware        []Middleware
	Target            *url.URL // TODO this is here just to allow the options to mutate it.
}

func New(opts ...Option) (*Transport, error) {

	t := &Transport{}
	if err := t.applyOptions(opts...); err != nil {
		return nil, err
	}

	if t.Requester == nil {
		client := nethttp.DefaultClient
		t.Requester = http.NewRequester(client, t.Target)
	}

	if t.Receiver == nil {
		t.Receiver = http.NewReceiver()
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
	switch t.Encoding {
	case Default, Binary:
		ctx = binding.WithForceBinary(ctx)
	case Structured:
		ctx = binding.WithForceStructured(ctx)
	}

	return t.BindingTransport.Send(ctx, e)
}

// StartReceiver implements Transport.StartReceiver
// NOTE: This is a blocking call.
func (t *Transport) StartReceiver(ctx context.Context) error {
	t.reMu.Lock()
	defer t.reMu.Unlock()

	if t.Handler == nil {
		t.Handler = nethttp.NewServeMux()
	}

	if !t.handlerRegistered {
		// handler.Handle might panic if the user tries to use the same path as the sdk.
		t.Handler.Handle(t.GetPath(), t)
		t.handlerRegistered = true
	}

	addr, err := t.listen()
	if err != nil {
		return err
	}

	t.server = &nethttp.Server{
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

func (t *Transport) ServeHTTP(rw nethttp.ResponseWriter, req *nethttp.Request) {
	if s, ok := t.BindingTransport.Receiver.(*http.Receiver); ok {
		s.ServeHTTP(rw, req)
	}
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

// attachMiddleware attaches the HTTP middleware to the specified handler.
func attachMiddleware(h nethttp.Handler, middleware []Middleware) nethttp.Handler {
	for _, m := range middleware {
		h = m(h)
	}
	return h
}
