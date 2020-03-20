package amqp

import (
	"bytes"
	"context"
	"reflect"
	"strings"

	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
)

const prefix = "cloudEvents:" // Name prefix for AMQP properties that hold CE attributes.

var (
	// Use the package path as AMQP error condition name
	condition = amqp.ErrorCondition(reflect.TypeOf(Message{}).PkgPath())
	specs     = spec.WithPrefix(prefix)
)

// Message implements binding.Message by wrapping an *amqp.Message.
// This message *can* be read several times safely
type Message struct {
	AMQP     *amqp.Message
	encoding binding.Encoding
}

// Wrap an *amqp.Message in a binding.Message.
// The returned message *can* be read several times safely
func NewMessage(message *amqp.Message) *Message {
	if message.Properties != nil && format.IsFormat(message.Properties.ContentType) {
		return &Message{AMQP: message, encoding: binding.EncodingStructured}
	} else if sv := getSpecVersion(message); sv != nil {
		return &Message{AMQP: message, encoding: binding.EncodingBinary}
	} else {
		return &Message{AMQP: message, encoding: binding.EncodingUnknown}
	}
}

var _ binding.Message = (*Message)(nil)

func getSpecVersion(message *amqp.Message) spec.Version {
	if sv, ok := message.ApplicationProperties[specs.PrefixedSpecVersionName()]; ok {
		if svs, ok := sv.(string); ok {
			return specs.Version(svs)
		}
	}
	return nil
}

func (m *Message) ReadEncoding() binding.Encoding {
	return m.encoding
}

func (m *Message) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	if m.AMQP.Properties != nil && format.IsFormat(m.AMQP.Properties.ContentType) {
		return encoder.SetStructuredEvent(ctx, format.Lookup(m.AMQP.Properties.ContentType), bytes.NewReader(m.AMQP.GetData()))
	}
	return binding.ErrNotStructured
}

func (m *Message) ReadBinary(ctx context.Context, encoder binding.BinaryWriter) error {
	version := getSpecVersion(m.AMQP)
	if version == nil {
		return binding.ErrNotBinary
	}

	err := encoder.Start(ctx)
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
	return encoder.End(ctx)
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
