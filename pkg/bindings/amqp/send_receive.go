package amqp

import (
	"context"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"pack.ag/amqp"
)

// Sender wraps an amqp.Sender as a binding.Sender
type Sender struct{ AMQP *amqp.Sender }

func (s Sender) Send(ctx context.Context, in binding.Message) error {
	var out *amqp.Message
	if m, ok := in.(Message); ok { // Already an AMQP message.
		out = m.AMQP
	} else if f, b := in.Structured(); f != "" { // Re-package structured message
		out = NewStruct(f, b)
	} else { // Construct a binary message
		e, err := in.Event()
		if err != nil {
			return err
		}
		out, err = NewBinary(e)
		if err != nil {
			return err
		}
	}
	err := s.AMQP.Send(ctx, out)
	_ = in.Finish(err)
	return err
}

func (s Sender) Close(ctx context.Context) error { return s.AMQP.Close(ctx) }

// Receiver wraps an amqp.Receiver as a binding.Receiver
type Receiver struct{ AMQP *amqp.Receiver }

func (r Receiver) Receive(ctx context.Context) (binding.Message, error) {
	m, err := r.AMQP.Receive(ctx)
	return Message{AMQP: m}, err
}

func (r Receiver) Close(ctx context.Context) error { return r.AMQP.Close(ctx) }
