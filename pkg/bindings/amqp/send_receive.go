package amqp

import (
	"context"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"pack.ag/amqp"
)

// Sender wraps an amqp.Sender as a binding.Sender
type Sender struct{ AMQP *amqp.Sender }

func (s Sender) Send(ctx context.Context, in binding.Message) (err error) {
	defer func() { _ = in.Finish(err) }()
	if m, ok := in.(Message); ok { // Already an AMQP message.
		return s.AMQP.Send(ctx, m.AMQP)
	}
	m, err := binding.Translate(in,
		BinaryEncoder{}.Encode,
		func(f string, b []byte) (binding.Message, error) { return Message{NewStruct(f, b)}, nil },
	)
	return s.AMQP.Send(ctx, m.(Message).AMQP)
}

func (s Sender) Close(ctx context.Context) error { return s.AMQP.Close(ctx) }

// Receiver wraps an amqp.Receiver as a binding.Receiver
type Receiver struct{ AMQP *amqp.Receiver }

func (r Receiver) Receive(ctx context.Context) (binding.Message, error) {
	m, err := r.AMQP.Receive(ctx)
	return Message{AMQP: m}, err
}

func (r Receiver) Close(ctx context.Context) error { return r.AMQP.Close(ctx) }
