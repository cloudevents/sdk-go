package amqp

import (
	"context"

	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/pkg/binding"
)

// sender wraps an amqp.Sender as a binding.Sender
type sender struct {
	amqp         *amqp.Sender
	transformers binding.TransformerFactories
}

func (s *sender) Send(ctx context.Context, in binding.Message) error {
	var err error
	defer func() { _ = in.Finish(err) }()
	if m, ok := in.(*Message); ok { // Already an AMQP message.
		return s.amqp.Send(ctx, m.AMQP)
	}

	var amqpMessage amqp.Message
	err = EncodeAMQPMessage(ctx, in, &amqpMessage, s.transformers)
	if err != nil {
		return err
	}

	return s.amqp.Send(ctx, &amqpMessage)
}

func (s *sender) Close(ctx context.Context) error { return s.amqp.Close(ctx) }

// Create a new Sender which wraps an amqp.Sender in a binding.Sender
func NewSender(amqpSender *amqp.Sender, options ...SenderOptionFunc) binding.Sender {
	s := &sender{amqp: amqpSender, transformers: make(binding.TransformerFactories, 0)}
	for _, o := range options {
		o(s)
	}
	return s
}
