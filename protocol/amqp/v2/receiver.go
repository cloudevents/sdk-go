package amqp

import (
	"context"
	"io"

	"github.com/Azure/go-amqp"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

// receiver wraps an amqp.Receiver as a binding.Receiver
type receiver struct{ amqp *amqp.Receiver }

func (r *receiver) Receive(ctx context.Context) (binding.Message, error) {
	var msg binding.Message

	if err := r.amqp.HandleMessage(ctx, func(m *amqp.Message) error {
		msg = NewMessage(m)
		return nil
	}); err != nil {
		if err == ctx.Err() {
			return nil, io.EOF
		}
		return nil, err
	}
	return msg, nil
}

// NewReceiver create a new Receiver which wraps an amqp.Receiver in a binding.Receiver
func NewReceiver(amqp *amqp.Receiver) protocol.Receiver {
	return &receiver{amqp: amqp}
}
