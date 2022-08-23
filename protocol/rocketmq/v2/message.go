package rocketmq

import (
	"bytes"
	"context"

	"github.com/apache/rocketmq-client-go/v2/primitive"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
)

// Message implements binding.Message by wrapping an *primitive.Message.
// This message *can* be read several times safely
type Message struct {
	Msg *primitive.Message
}

// NewMessage wraps a *primitive.Message in a binding.Message.
// The returned message *can* be read several times safely
func NewMessageFromConsumerMessage(msg *primitive.Message) *Message {
	return &Message{Msg: msg}
}

var _ binding.Message = (*Message)(nil)

func (m *Message) ReadEncoding() binding.Encoding {
	return binding.EncodingStructured
}

func (m *Message) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	return encoder.SetStructuredEvent(ctx, format.JSON, bytes.NewReader(m.Msg.body))
}

func (m *Message) ReadBinary(context.Context, binding.BinaryWriter) error {
	return binding.ErrNotBinary
}

func (m *Message) Finish(err error) error {
	return nil
}
