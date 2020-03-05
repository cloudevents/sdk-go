package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	nethttp "net/http"

	"github.com/cloudevents/sdk-go/pkg/binding"
)

type msgErr struct {
	msg *Message
	err error
}

// Receiver for CloudEvents as HTTP requests which implements nethttp.Handler.
// To receive messages, associate it with a nethttp.Server.
type Receiver struct {
	incoming chan msgErr
}

// ServeHTTP implements nethttp.Handler.
// Blocks until Message.Finish is called.
func (r *Receiver) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var err error
	m := NewMessageFromHttpRequest(req)
	if m.Encoding() == binding.EncodingUnknown {
		r.incoming <- msgErr{nil, binding.ErrUnknownEncoding}
	}
	done := make(chan error)
	m.onFinish = func(err error) error { done <- err; return nil }
	r.incoming <- msgErr{m, err} // Send to Receive()
	if err = <-done; err != nil {
		nethttp.Error(rw, fmt.Sprintf("cannot forward CloudEvent: %v", err), http.StatusInternalServerError)
	}
}

// NewReceiver creates a Receiver which implements nethttp.Handler.
// To receive messages, associate it with a nethttp.Server.
func NewReceiver() *Receiver {
	return &Receiver{incoming: make(chan msgErr)}
}

// Receive the next incoming HTTP request as a CloudEvent.
// Returns non-nil error if the incoming HTTP request fails to parse as a CloudEvent
// Returns io.EOF if the receiver is closed.
func (r *Receiver) Receive(ctx context.Context) (binding.Message, error) {
	msgErr, ok := <-r.incoming
	if !ok {
		return nil, io.EOF
	}
	return msgErr.msg, msgErr.err
}
