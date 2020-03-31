package pubsub

import (
	"bytes"
	"context"
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
	internal *pubsub.Message
	format   format.Format
	version  spec.Version
}

// NewMessage returns a binding.Message with data and attributes.
// This message *can* be read several times safely
func NewMessage(pm *pubsub.Message) *Message {
	var f format.Format = nil
	var version spec.Version = nil
	if pm.Attributes != nil {
		// Use Content-type attr to determine if message is structured and
		// set format.
		if s := pm.Attributes[contentType]; format.IsFormat(s) {
			f = format.Lookup(s)
		}
		// Binary v0.3:
		if s := pm.Attributes[specs.PrefixedSpecVersionName()]; s != "" {
			version = specs.Version(s)
		}
	}

	return &Message{
		internal: pm,
		format:   f,
		version:  version,
	}
}

// Check if pubsub.Message implements binding.Message
var _ binding.Message = (*Message)(nil)

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
	if m.version != nil {
		return binding.ErrNotStructured
	}
	return encoder.SetStructuredEvent(ctx, m.format, bytes.NewReader(m.internal.Data))
}

func (m *Message) ReadBinary(ctx context.Context, encoder binding.BinaryWriter) error {
	if m.format != nil {
		return binding.ErrNotBinary
	}

	err := encoder.Start(ctx)
	if err != nil {
		return err
	}

	for k, v := range m.internal.Attributes {
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

	err = encoder.SetData(bytes.NewReader(m.internal.Data))
	if err != nil {
		return err
	}

	return encoder.End(ctx)
}

// Finish marks the message to be forgotten.
// If err is nil, the underlying Psubsub message will be acked;
// otherwise nacked.
func (m *Message) Finish(err error) error {
	if err != nil {
		m.internal.Nack()
	} else {
		m.internal.Ack()
	}
	return nil
}
