package transport

import (
	"context"

	"github.com/cloudevents/sdk-go/pkg/binding"
)

// ResponseFn is the function callback provided from Responder.Respond to allow
// for a receiver to "reply" to a message it receives.
type ResponseFn func(ctx context.Context, m binding.Message) error

// Responder receives messages and is given a callback to respond.
type Responder interface {
	// Receive blocks till a message is received or ctx expires.
	//
	// A non-nil error means the receiver is closed.
	// io.EOF means it closed cleanly, any other value indicates an error.
	Respond(ctx context.Context) (binding.Message, ResponseFn, error)
}

// ResponderCloser is a Responder that can be closed.
type ResponderCloser interface {
	Responder
	Closer
}
