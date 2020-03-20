package amqp

import (
	"context"
	"io"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"pack.ag/amqp"
)

// receiver wraps an amqp.Receiver as a binding.Receiver
type receiver struct{ amqp *amqp.Receiver }

func (r *receiver) Receive(ctx context.Context) (binding.Message, error) {
	m, err := r.amqp.Receive(ctx)
	if err != nil {
		if err == ctx.Err() {
			return nil, io.EOF
		}
		return nil, err
	}

	return NewMessage(m), nil
}

func (r *receiver) Close(ctx context.Context) error { return r.amqp.Close(ctx) }

// Create a new Receiver which wraps an amqp.Receiver in a binding.Receiver
func NewReceiver(amqp *amqp.Receiver) protocol.Receiver {
	return &receiver{amqp: amqp}
}
