package kafka_sarama

import (
	"bytes"
	"context"
	"strings"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"

	"github.com/Shopify/sarama"
)

const prefix = "ce_"

var specs = spec.WithPrefix(prefix)

const ContentType = "content-type"

// Message holds a sarama ConsumerMessage.
type Message struct {
	key, value  []byte
	headers     map[string][]byte
	contentType string
	format      format.Format
	version     spec.Version
}

// Check if http.Message implements binding.Message
var _ binding.Message = (*Message)(nil)

// NewMessage returns a Message with data from body.
// Reads and closes body.
func NewMessage(cm *sarama.ConsumerMessage) (*Message, error) {
	var contentType string
	headers := make(map[string][]byte, len(cm.Headers))
	for _, r := range cm.Headers {
		k := strings.ToLower(string(r.Key))
		if k == ContentType {
			contentType = string(r.Value)
		}
		headers[strings.ToLower(string(r.Key))] = r.Value
	}
	return NewMessageFromRaw(cm.Key, cm.Value, contentType, headers)
}

func NewMessageFromRaw(key []byte, value []byte, contentType string, headers map[string][]byte) (*Message, error) {
	m := Message{
		key:         key,
		value:       value,
		contentType: contentType,
		headers:     headers,
	}
	if ft := format.Lookup(contentType); ft != nil {
		m.format = ft
	} else if v, err := specs.FindVersion(func(s string) string {
		return string(headers[s])
	}); err == nil {
		m.version = v
	}
	return &m, nil
}

func (m *Message) Encoding() binding.Encoding {
	if m.version != nil {
		return binding.EncodingBinary
	}
	if m.format != nil {
		return binding.EncodingStructured
	}
	return binding.EncodingUnknown
}

func (m *Message) Structured(ctx context.Context, encoder binding.StructuredEncoder) error {
	if m.format != nil {
		return encoder.SetStructuredEvent(ctx, m.format, bytes.NewReader(m.value))
	}
	return binding.ErrNotStructured
}

func (m *Message) Binary(ctx context.Context, encoder binding.BinaryEncoder) error {
	if m.version == nil {
		return binding.ErrNotBinary
	}

	err := encoder.Start(ctx)
	if err != nil {
		return err
	}

	for k, v := range m.headers {
		if strings.HasPrefix(k, prefix) {
			attr := m.version.Attribute(k)
			if attr != nil {
				err = encoder.SetAttribute(attr, string(v))
			} else {
				err = encoder.SetExtension(strings.ToLower(strings.TrimPrefix(k, prefix)), string(v))
			}
		} else if k == ContentType {
			err = encoder.SetAttribute(m.version.AttributeFromKind(spec.DataContentType), string(v))
		}
		if err != nil {
			return err
		}
	}

	if m.key != nil {
		err = encoder.SetExtension("key", string(m.key))
		if err != nil {
			return err
		}
	}

	if m.value != nil {
		err = encoder.SetData(bytes.NewReader(m.value))
		if err != nil {
			return err
		}
	}

	return encoder.End()
}

func (m *Message) Finish(error) error {
	return nil
}
