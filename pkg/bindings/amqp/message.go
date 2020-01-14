package amqp

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"

	"pack.ag/amqp"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
)

const prefix = "cloudEvents:" // Name prefix for AMQP properties that hold CE attributes.

var (
	// Use the package path as AMQP error condition name
	condition = amqp.ErrorCondition(reflect.TypeOf(Message{}).PkgPath())
	specs     = spec.WithPrefix(prefix)
)

// Message implements binding.Message by wrapping an *amqp.Message.
type Message struct{ AMQP *amqp.Message }

// Check if amqp.Message implements binding.Message & binding.MessagePayloadReader
var _ binding.Message = (*Message)(nil)
var _ binding.MessagePayloadReader = (*Message)(nil)

func (m *Message) Structured(encoder binding.StructuredEncoder) error {
	if m.AMQP.Properties != nil && format.IsFormat(m.AMQP.Properties.ContentType) {
		return encoder.SetStructuredEvent(format.Lookup(m.AMQP.Properties.ContentType), m)
	}
	return binding.ErrNotStructured
}

func (m *Message) Binary(encoder binding.BinaryEncoder) error {
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

	if !m.IsEmpty() {
		return encoder.SetData(m)
	}
	return nil
}

func (m *Message) Event(encoder binding.EventEncoder) error {
	e, _, _, err := binding.ToEvent(m)
	if err != nil {
		return err
	}
	return encoder.SetEvent(e)
}

func (m *Message) IsEmpty() bool {
	return m.AMQP.Data == nil || len(m.AMQP.Data) == 0
}

func (m *Message) Bytes() []byte {
	return m.AMQP.GetData()
}

func (m *Message) Reader() io.Reader {
	return bytes.NewReader(m.AMQP.GetData())
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
