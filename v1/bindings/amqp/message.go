package amqp

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"

	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/format"
	"github.com/cloudevents/sdk-go/v1/binding/spec"
)

const prefix = "cloudEvents:" // Name prefix for AMQP properties that hold CE attributes.

var (
	// Use the package path as AMQP error condition name
	condition = amqp.ErrorCondition(reflect.TypeOf(Message{}).PkgPath())
	specs     = spec.WithPrefix(prefix)
)

// Message implements binding.Message by wrapping an *amqp.Message.
type Message struct {
	AMQP     *amqp.Message
	encoding binding.Encoding
}

func NewMessage(message *amqp.Message) *Message {
	if message.Properties != nil && format.IsFormat(message.Properties.ContentType) {
		return &Message{AMQP: message, encoding: binding.EncodingStructured}
	} else if _, err := specs.FindVersion(func(k string) string {
		s, _ := message.ApplicationProperties[k].(string)
		return s
	}); err == nil {
		return &Message{AMQP: message, encoding: binding.EncodingBinary}
	} else {
		return &Message{AMQP: message, encoding: binding.EncodingUnknown}
	}
}

// Check if amqp.Message implements binding.Message
var _ binding.Message = (*Message)(nil)

func (m *Message) Encoding() binding.Encoding {
	return m.encoding
}

func (m *Message) Structured(ctx context.Context, encoder binding.StructuredEncoder) error {
	if m.AMQP.Properties != nil && format.IsFormat(m.AMQP.Properties.ContentType) {
		return encoder.SetStructuredEvent(ctx, format.Lookup(m.AMQP.Properties.ContentType), bytes.NewReader(m.AMQP.GetData()))
	}
	return binding.ErrNotStructured
}

func (m *Message) Binary(ctx context.Context, encoder binding.BinaryEncoder) error {
	if len(m.AMQP.ApplicationProperties) == 0 {
		return errors.New("AMQP CloudEvents message has no application properties")
	}
	version, err := specs.FindVersion(func(k string) string {
		s, _ := m.AMQP.ApplicationProperties[k].(string)
		return s
	})
	if err != nil {
		return err
	}

	err = encoder.Start(ctx)
	if err != nil {
		return err
	}

	if m.AMQP.Properties != nil && m.AMQP.Properties.ContentType != "" {
		err = encoder.SetAttribute(version.AttributeFromKind(spec.DataContentType), m.AMQP.Properties.ContentType)
		if err != nil {
			return err
		}
	}

	for k, v := range m.AMQP.ApplicationProperties {
		if strings.HasPrefix(k, prefix) {
			attr := version.Attribute(k)
			if attr != nil {
				err = encoder.SetAttribute(attr, v)
			} else {
				err = encoder.SetExtension(strings.ToLower(strings.TrimPrefix(k, prefix)), v)
			}
		}
		if err != nil {
			return err
		}
	}

	data := m.AMQP.GetData()
	if len(data) != 0 { // Some data
		err = encoder.SetData(bytes.NewReader(data))
		if err != nil {
			return err
		}
	}
	return encoder.End()
}

func (m *Message) Finish(err error) error {
	if err != nil {
		return m.AMQP.Reject(&amqp.Error{
			Condition:   condition,
			Description: err.Error(),
		})
	}
	return m.AMQP.Accept()
}
