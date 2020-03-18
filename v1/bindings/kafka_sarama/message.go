package kafka_sarama

import (
	"bytes"
	"context"
	"strings"

	ce "github.com/cloudevents/sdk-go/v1"
	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/format"
	"github.com/cloudevents/sdk-go/v1/binding/spec"

	"github.com/Shopify/sarama"
)

const prefix = "ce_"

var specs = spec.WithPrefix(prefix)

const ContentType = "content-type"

// Message holds a sarama ConsumerMessage.
type Message struct {
	Key, Value  []byte
	Headers     map[string][]byte
	ContentType string
	format      format.Format
	version     spec.Version
}

// Check if http.Message implements binding.Message
var _ binding.Message = (*Message)(nil)

// NewMessage returns a Message with data from body.
// Reads and closes body.
func NewMessage(cm *sarama.ConsumerMessage) (binding.Message, error) {
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

func NewMessageFromRaw(key []byte, value []byte, contentType string, headers map[string][]byte) (binding.Message, error) {
	if ft := format.Lookup(contentType); ft != nil {
		// if the message is structured and has a key,
		// then it's cheaper to go through event message
		// because we need to add the key as extension
		if key != nil {
			event := ce.Event{}
			err := ft.Unmarshal(value, &event)
			if err != nil {
				return nil, err
			}
			event.SetExtension("key", string(key))
			return binding.EventMessage(event), nil
		} else {
			return &Message{
				Key:         key,
				Value:       value,
				ContentType: contentType,
				Headers:     headers,
				format:      ft,
			}, nil
		}
	} else if v, err := specs.FindVersion(func(s string) string {
		return string(headers[strings.ToLower(s)])
	}); err == nil {
		return &Message{
			Key:         key,
			Value:       value,
			ContentType: contentType,
			Headers:     headers,
			version:     v,
		}, nil
	}

	return &Message{
		Key:         key,
		Value:       value,
		ContentType: contentType,
		Headers:     headers,
	}, nil
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
		return encoder.SetStructuredEvent(ctx, m.format, bytes.NewReader(m.Value))
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

	for k, v := range m.Headers {
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

	if m.Key != nil {
		err = encoder.SetExtension("key", string(m.Key))
		if err != nil {
			return err
		}
	}

	if m.Value != nil {
		err = encoder.SetData(bytes.NewReader(m.Value))
		if err != nil {
			return err
		}
	}

	return encoder.End()
}

func (m *Message) Finish(error) error {
	return nil
}
