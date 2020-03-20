package amqp

import (
	"context"

	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
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
		err = s.amqp.Send(ctx, m.AMQP)
		return err
	}

	var amqpMessage amqp.Message
	err = WriteMessage(ctx, in, &amqpMessage, s.transformers)
	if err != nil {
		return err
	}

	err = s.amqp.Send(ctx, &amqpMessage)
	return err
}

func (s *sender) Close(ctx context.Context) error { return s.amqp.Close(ctx) }

// Create a new Sender which wraps an amqp.Sender in a binding.Sender
func NewSender(amqpSender *amqp.Sender, options ...SenderOptionFunc) protocol.Sender {
	s := &sender{amqp: amqpSender, transformers: make(binding.TransformerFactories, 0)}
	for _, o := range options {
		o(s)
	}
	return s
}
