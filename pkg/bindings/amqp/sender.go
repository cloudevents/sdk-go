package amqp

import (
	"context"

	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
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
	err = s.fillAMQPRequest(&amqpMessage, in)
	if err != nil {
		return err
	}

	return s.AMQP.Send(ctx, &amqpMessage)
}

// This function tries:
// 1. Translate from structured
// 2. Translate from binary
// 3. Translate to Event and then re-encode back to amqp.Message
func (s *Sender) fillAMQPRequest(amqpMessage *amqp.Message, m binding.Message) error {
	createStructured := func() binding.StructuredEncoder {
		return &structuredMessageEncoder{amqpMessage}
	}
	if s.forceBinary {
		createStructured = nil
	}

	createBinary := func() binding.BinaryEncoder {
		return newBinaryMessageEncoder(amqpMessage)
	}
	if s.forceStructured {
		createBinary = nil
	}

	createEvent := func() binding.EventEncoder {
		if s.forceStructured {
			return &eventToStructuredMessageEncoder{format: format.JSON, amqpMessage: amqpMessage}
		}
		return &eventToBinaryMessageEncoder{amqpMessage}
	}

	_, _, err := binding.Translate(m, createStructured, createBinary, createEvent, s.transformerFactories)
	return err
}

func (s *Sender) Close(ctx context.Context) error { return s.AMQP.Close(ctx) }

func NewSender(amqpClient *amqp.Sender, options ...SenderOptionFunc) binding.Sender {
	s := &Sender{AMQP: amqpClient, transformerFactories: make(binding.TransformerFactories, 0)}
	for _, o := range options {
		o(s)
	}
	return s
}
