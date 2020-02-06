package amqp

import (
	"context"

	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/pkg/binding"
)

// Sender wraps an amqp.Sender as a binding.Sender
type Sender struct {
	AMQP *amqp.Sender

	transformerFactories binding.TransformerFactories

	forceBinary     bool
	forceStructured bool
}

func (s *Sender) Send(ctx context.Context, in binding.Message) error {
	var err error
	defer func() { _ = in.Finish(err) }()
	if m, ok := in.(*Message); ok { // Already an AMQP message.
		return s.AMQP.Send(ctx, m.AMQP)
	}

	var amqpMessage amqp.Message
	err = EncodeAMQPMessage(in, &amqpMessage, s.forceStructured, s.forceBinary, s.transformerFactories)
	if err != nil {
		return err
	}

	return s.AMQP.Send(ctx, &amqpMessage)
}

func (s *Sender) Close(ctx context.Context) error { return s.AMQP.Close(ctx) }

func NewSender(amqpClient *amqp.Sender, options ...SenderOptionFunc) binding.Sender {
	s := &Sender{AMQP: amqpClient, transformerFactories: make(binding.TransformerFactories, 0)}
	for _, o := range options {
		o(s)
	}
	return s
}
