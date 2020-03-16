package transport

import (
	"context"

	"github.com/cloudevents/sdk-go/pkg/binding"
)

// Requester sends a message and receives a response
//
// Optional interface that may be implemented by protocols that support
// request/response correlation.
type Requester interface {
	// Request sends m like Sender.Send() but also arranges to receive a response.
	Request(ctx context.Context, m binding.Message) (binding.Message, error)
}

// RequesterCloser is a Requester that can be closed.
type RequesterCloser interface {
	Requester
	Closer
}
