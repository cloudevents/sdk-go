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

	transcoderFactories binding.TranscoderFactories

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
	if !s.forceBinary {
		b := s.transcoderFactories.StructuredMessageTranscoder(&structuredMessageBuilder{amqpMessage})
		if b != nil {
			if err := m.Structured(b); err == nil {
				return nil
			} else if err != binding.ErrNotStructured {
				return err
			}
		}
	}

	if !s.forceStructured {
		b := s.transcoderFactories.BinaryMessageTranscoder(newBinaryMessageBuilder(amqpMessage))
		if b != nil {
			if err := m.Binary(b); err == nil {
				return nil
			} else if err != binding.ErrNotBinary {
				return err
			}
		}
	}

	if s.forceStructured {
		return m.Event(
			s.transcoderFactories.EventMessageTranscoder(&eventToStructuredMessageBuilder{format: format.JSON, amqpMessage: amqpMessage}),
		)
	} else {
		return m.Event(
			s.transcoderFactories.EventMessageTranscoder(&eventToBinaryMessageBuilder{amqpMessage}),
		)
	}
}

func (s *Sender) Close(ctx context.Context) error { return s.AMQP.Close(ctx) }

func NewSender(amqpClient *amqp.Sender, options ...SenderOptionFunc) binding.Sender {
	s := &Sender{AMQP: amqpClient, transcoderFactories: make(binding.TranscoderFactories, 0)}
	for _, o := range options {
		o(s)
	}
	return s
}
