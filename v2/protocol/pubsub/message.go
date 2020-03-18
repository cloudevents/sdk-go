package pubsub

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
)

const (
	prefix      = "ce-"
	contentType = "Content-Type"
)

var specs = spec.WithPrefix(prefix)

// Message represents a Pub/Sub message.
// This message *can* be read several times safely
type Message struct {
	Msg pubsub.Message

	format   format.Format
	version  spec.Version
	encoding binding.Encoding
}

// NewMessage returns a binding.Message with data and attributes.
// This message *can* be read several times safely
func NewMessage(data []byte, attributes map[string]string) *Message {
	encoding := binding.EncodingBinary
	var f format.Format = nil
	var version spec.Version = nil
	if attributes != nil {
		// Use Content-type attr to determine if message is structured and
		// set format.
		if s := attributes[contentType]; format.IsFormat(s) {
			f = format.Lookup(s)
			encoding = binding.EncodingStructured
		}
		// Binary v0.3:
		if s := attributes[prefix+"specversion"]; s != "" {
			version = specs.Version(s)
		}
	}

	// Check as Structured encoding.
	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &raw); err != nil {
		version = specs.Version("")
	}

	// Structured v0.3
	if v, ok := raw["specversion"]; ok {
		var ver string
		if err := json.Unmarshal(v, &ver); err != nil {
			version = specs.Version("")
		}
		version = specs.Version(ver)
	}
	return &Message{
		Msg:      pubsub.Message{Data: data, Attributes: attributes},
		encoding: encoding,
		format:   f,
		version:  version,
	}
}

// Check if pubsub.Message implements binding.Message
var _ binding.Message = (*Message)(nil)

func (m *Message) ReadEncoding() binding.Encoding {
	return m.encoding
}

func (m *Message) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	if m.encoding == binding.EncodingBinary {
		return binding.ErrNotStructured
	}
	return encoder.SetStructuredEvent(ctx, m.format, bytes.NewReader(m.Msg.Data))
}

func (m *Message) ReadBinary(ctx context.Context, encoder binding.BinaryWriter) error {
	if m.encoding == binding.EncodingStructured {
		return binding.ErrNotBinary
	}

	err := encoder.Start(ctx)
	if err != nil {
		return err
	}

	for k, v := range m.Msg.Attributes {
		if strings.HasPrefix(k, prefix) {
			attr := m.version.Attribute(k)
			if attr != nil {
				err = encoder.SetAttribute(attr, string(v))
			} else {
				err = encoder.SetExtension(strings.TrimPrefix(k, prefix), string(v))
			}
		} else if k == contentType {
			err = encoder.SetAttribute(m.version.AttributeFromKind(spec.DataContentType), string(v))
		}
		if err != nil {
			return err
		}
	}

	err = encoder.SetData(bytes.NewReader(m.Msg.Data))
	if err != nil {
		return err
	}

	return encoder.End(ctx)
}

func (m *Message) Finish(err error) error {
	return nil
}
