package amqp

import (
	"context"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"pack.ag/amqp"
)

// Receiver wraps an amqp.Receiver as a binding.Receiver
type Receiver struct{ AMQP *amqp.Receiver }

func (r *Receiver) Receive(ctx context.Context) (binding.Message, error) {
	m, err := r.AMQP.Receive(ctx)
	return NewMessage(m), err
}

func (r *Receiver) Close(ctx context.Context) error { return r.AMQP.Close(ctx) }
