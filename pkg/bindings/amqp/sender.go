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
	amqpMessage.Properties = &amqp.MessageProperties{}
	amqpMessage.ApplicationProperties = make(map[string]interface{})

	var structuredEncoder binding.StructuredEncoder
	if !s.forceBinary {
		structuredEncoder = (*amqpMessageEncoder)(amqpMessage)
	}

	var binaryEncoder binding.BinaryEncoder
	if !s.forceStructured {
		binaryEncoder = (*amqpMessageEncoder)(amqpMessage)
	}

	var preferredEventEncoding binding.Encoding
	if s.forceStructured {
		preferredEventEncoding = binding.EncodingStructured
	} else {
		preferredEventEncoding = binding.EncodingBinary
	}

	_, err := binding.Encode(
		m,
		structuredEncoder,
		binaryEncoder,
		s.transformerFactories,
		preferredEventEncoding,
	)
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
