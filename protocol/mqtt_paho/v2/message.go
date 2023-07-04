package mqtt_paho

import (
	"bytes"
	"context"
	"strings"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/eclipse/paho.golang/paho"
)

const (
	prefix      = "ce-"
	contentType = "Content-Type"
)

var specs = spec.WithPrefix(prefix)

// Message represents a MQTT message.
// This message *can* be read several times safely
type Message struct {
	internal *paho.Publish
	version  spec.Version
	format   format.Format
}

// Check if .Message implements binding.Message
var (
	_ binding.Message               = (*Message)(nil)
	_ binding.MessageMetadataReader = (*Message)(nil)
)

func NewMessage(msg *paho.Publish) *Message {
	var f format.Format = nil
	var v spec.Version = nil
	if msg.Properties != nil {
		// Use Content-type attr to determine if message is structured and
		// set format.
		if s := msg.Properties.ContentType; format.IsFormat(s) {
			f = format.Lookup(s)
		}
		// Binary v0.3:
		if s := msg.Properties.User.Get(specs.PrefixedSpecVersionName()); s != "" {
			v = specs.Version(s)
		}
	}
	return &Message{
		internal: msg,
		version:  v,
		format:   f,
	}
}

func (m *Message) ReadEncoding() binding.Encoding {
	if m.version != nil {
		return binding.EncodingBinary
	}
	if m.format != nil {
		return binding.EncodingStructured
	}
	return binding.EncodingUnknown
}

func (m *Message) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	if m.format != nil {
		return encoder.SetStructuredEvent(ctx, m.format, bytes.NewReader(m.internal.Payload))
	}
	return binding.ErrNotStructured
}

func (m *Message) ReadBinary(ctx context.Context, encoder binding.BinaryWriter) (err error) {
	if m.version == nil {
		return binding.ErrNotBinary
	}
	if v := m.internal.Properties.ContentType; v != "" {
		err = encoder.SetAttribute(m.version.AttributeFromKind(spec.DataContentType), v)
		if err != nil {
			return err
		}
	}

	for _, userProperty := range m.internal.Properties.User {
		if strings.HasPrefix(userProperty.Key, prefix) {
			attr := m.version.Attribute(userProperty.Key)
			if attr != nil {
				err = encoder.SetAttribute(attr, userProperty.Value)
			} else {
				err = encoder.SetExtension(strings.TrimPrefix(userProperty.Key, prefix), userProperty.Value)
			}
		}
		if err != nil {
			return
		}
	}

	if m.internal.Payload != nil {
		return encoder.SetData(bytes.NewBuffer(m.internal.Payload))
	}
	return nil
}

func (m *Message) Finish(error) error {
	return nil
}

func (m *Message) GetAttribute(k spec.Kind) (spec.Attribute, interface{}) {
	if k == spec.DataContentType {
		return m.version.AttributeFromKind(spec.DataContentType), m.internal.Properties.ContentType
	}

	attr := m.version.AttributeFromKind(k)
	if attr != nil {
		return attr, m.internal.Properties.User.Get(prefix + attr.Name())
	}
	return nil, nil
}

func (m *Message) GetExtension(name string) interface{} {
	return m.internal.Properties.User.Get(prefix + name)
}
