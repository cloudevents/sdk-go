package fasthttp

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"net"
	"net/http"
	"strings"

	"github.com/cloudevents/sdk-go/v2/protocol"
)

var _ protocol.Opener = (*Protocol)(nil)

const compress = false

func (p *Protocol) OpenInbound(ctx context.Context) error {
	p.reMu.Lock()
	defer p.reMu.Unlock()

	// TODO: support handlers.
	//if p.Handler == nil {
	//	p.Handler = http.NewServeMux()
	//}

	//if !p.handlerRegistered {
	//	// handler.Handle might panic if the user tries to use the same path as the sdk.
	//	p.Handler.Handle(p.GetPath(), p)
	//	p.handlerRegistered = true
	//}

	// After listener is invoked
	listener, err := p.listen()
	if err != nil {
		return err
	}

	h := p.requestHandler
	if compress {
		h = fasthttp.CompressHandler(h)
	}

	p.server = &fasthttp.Server{
		Handler: h,
		// TODO: support trace handlers
		//Handler: &ochttp.Handler{
		//	Propagation:    &tracecontext.HTTPFormat{},
		//	Handler:        attachMiddleware(p.Handler, p.middleware),
		//	FormatSpanName: formatSpanName,
		//},
	}

	// Shutdown
	defer func() {
		_ = p.server.Shutdown()
		p.server = nil
	}()

	errChan := make(chan error, 1)
	go func() {
		errChan <- p.server.Serve(listener)
	}()

	// wait for the server to return or ctx.Done().
	select {
	case <-ctx.Done():
		// Try a gracefully shutdown.
		p.server.ReadTimeout = p.ShutdownTimeout
		err := p.server.Shutdown()
		<-errChan // Wait for server goroutine to exit
		return err
	case err := <-errChan:
		return err
	}
}

// GetListeningPort returns the listening port.
// Returns -1 if it's not listening.
func (p *Protocol) GetListeningPort() int {
	if listener := p.listener.Load(); listener != nil {
		if tcpAddr, ok := listener.(net.Listener).Addr().(*net.TCPAddr); ok {
			return tcpAddr.Port
		}
	}
	return -1
}

func formatSpanName(r *http.Request) string {
	return "cloudevents.http." + r.URL.Path
}

// listen if not already listening, update t.Port
func (p *Protocol) listen() (net.Listener, error) {
	if p.listener.Load() == nil {
		port := 8080
		if p.Port != -1 {
			port = p.Port
			if port < 0 || port > 65535 {
				return nil, fmt.Errorf("invalid port %d", port)
			}
		}
		var err error
		var listener net.Listener
		if listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
			return nil, err
		}
		p.listener.Store(listener)
		return listener, nil
	}
	return p.listener.Load().(net.Listener), nil
}

// GetPath returns the path the transport is hosted on. If the path is '/',
// the transport will handle requests on any URI. To discover the true path
// a request was received on, inspect the context from Receive(cxt, ...) with
// TransportContextFrom(ctx).
func (p *Protocol) GetPath() string {
	path := strings.TrimSpace(p.Path)
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
