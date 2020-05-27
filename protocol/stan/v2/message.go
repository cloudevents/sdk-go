package stan

import (
	"bytes"
	"context"
	"github.com/nats-io/stan.go"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
)

// Message implements binding.Message by wrapping an *stan.Msg.
// This message *can* be read several times safely
type Message struct {
	Msg        *stan.Msg
	manualAcks bool
}

// NewMessage wraps a *nats.Msg in a binding.Message.
// The returned message *can* be read several times safely
func NewMessage(msg *stan.Msg, opts ...MessageOption) (*Message, error) {
	m := &Message{Msg: msg}

	err := m.applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Message) applyOptions(opts ...MessageOption) error {
	for _, fn := range opts {
		if err := fn(m); err != nil {
			return err
		}
	}
	return nil
}

var _ binding.Message = (*Message)(nil)

func (m *Message) ReadEncoding() binding.Encoding {
	return binding.EncodingStructured
}

func (m *Message) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	return encoder.SetStructuredEvent(ctx, format.JSON, bytes.NewReader(m.Msg.Data))
}

func (m *Message) ReadBinary(context.Context, binding.BinaryWriter) error {
	return binding.ErrNotBinary
}

func (m *Message) Finish(err error) error {
	if m.manualAcks && err == nil {
		return m.Msg.Ack()
	}

	return nil
}
