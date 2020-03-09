package transport

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/event"
)

// Transport is the interface for transport sender to send the converted Message
// over the underlying transport.
type Transport interface {
	Send(context.Context, event.Event) error
	Request(context.Context, event.Event) (*event.Event, error)

	SetDelivery(Delivery)
	StartReceiver(context.Context) error

	// HasTracePropagation is true when the transport implements
	// in-band trace propagation. When false, the client receiver
	// will propagate trace context from distributed tracing
	// extension attributes when available.
	HasTracePropagation() bool
}

// REMOVE:
// Receiver is an interface to define how a transport will invoke a listener
// of incoming events.
// Deprecated: use bindings.
type Delivery interface {
	// Deprecated: use bindings.
	Delivery(context.Context, event.Event, *event.EventResponse) error
}

// DeliveryFunc wraps a function as a Receiver object.
type DeliveryFunc func(ctx context.Context, e event.Event, er *event.EventResponse) error

// Receive implements Receiver.Receive
func (f DeliveryFunc) Receive(ctx context.Context, e event.Event, er *event.EventResponse) error {
	return f(ctx, e, er)
}
