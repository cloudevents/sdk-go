package kafka_sarama

import (
	"bytes"
	"context"
	"io"

	"github.com/Shopify/sarama"

	ce "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/types"
)

const (
	SKIP_KEY_EXTENSION = "SKIP_KEY_EXTENSION"
)

// Skip the key extension while encoding to Kafka ProducerMessage
func WithSkipKeyExtension(ctx context.Context) context.Context {
	return context.WithValue(ctx, SKIP_KEY_EXTENSION, true)
}

// Fill the provided producerMessage with the message m.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
// You can skip the key extension handling decorating the context using WithSkipKeyExtension:
// https://github.com/cloudevents/spec/blob/master/kafka-protocol-binding.md#31-key-attribute
func WriteKafkaProducerMessage(ctx context.Context, m binding.Message, producerMessage *sarama.ProducerMessage, transformerFactories binding.TransformerFactories) error {
	skipKey := binding.GetOrDefaultFromCtx(ctx, SKIP_KEY_EXTENSION, false).(bool)

	if skipKey {
		enc := &kafkaProducerMessageWriter{
			producerMessage,
			skipKey,
		}

		_, err := binding.Write(
			ctx,
			m,
			enc,
			enc,
			transformerFactories,
		)
		return err
	}

	enc := m.ReadEncoding()
	var err error
	// Skip direct encoding if the event is an event message
	if enc == binding.EncodingBinary {
		encoder := &kafkaProducerMessageWriter{
			producerMessage,
			skipKey,
		}
		enc, err = binding.DirectWrite(ctx, m, nil, encoder, transformerFactories)
		if enc != binding.EncodingUnknown {
			// Message directly encoded binary -> binary, nothing else to do here
			return err
		}
	}

	var e *ce.Event
	e, err = binding.ToEvent(ctx, m, transformerFactories)
	if err != nil {
		return err
	}

	if val, ok := e.Extensions()["key"]; ok {
		s, err := types.Format(val)
		if err != nil {
			return err
		}

		producerMessage.Key = sarama.StringEncoder(s)
	}

	eventMessage := binding.EventMessage(*e)

	encoder := &kafkaProducerMessageWriter{
		producerMessage,
		skipKey,
	}

	if binding.GetOrDefaultFromCtx(ctx, binding.PREFERRED_EVENT_ENCODING, binding.EncodingBinary).(binding.Encoding) == binding.EncodingStructured {
		return eventMessage.ReadStructured(ctx, encoder)
	} else {
		return eventMessage.ReadBinary(ctx, encoder)
	}
}

type kafkaProducerMessageWriter struct {
	*sarama.ProducerMessage
	skipKey bool
}

func (b *kafkaProducerMessageWriter) SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) error {
	b.Headers = []sarama.RecordHeader{{
		Key:   []byte(contentTypeHeader),
		Value: []byte(format.MediaType()),
	}}

	var buf bytes.Buffer
	_, err := io.Copy(&buf, event)
	if err != nil {
		return err
	}

	b.Value = sarama.ByteEncoder(buf.Bytes())
	return nil
}

func (b *kafkaProducerMessageWriter) Start(ctx context.Context) error {
	b.Headers = []sarama.RecordHeader{}
	return nil
}

func (b *kafkaProducerMessageWriter) End() error {
	return nil
}

func (b *kafkaProducerMessageWriter) SetData(reader io.Reader) error {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, reader)
	if err != nil {
		return err
	}

	b.Value = sarama.ByteEncoder(buf.Bytes())
	return nil
}

func (b *kafkaProducerMessageWriter) SetAttribute(attribute spec.Attribute, value interface{}) error {
	// Everything is a string here
	s, err := types.Format(value)
	if err != nil {
		return err
	}

	if attribute.Kind() == spec.DataContentType {
		b.Headers = append(b.Headers, sarama.RecordHeader{Key: []byte(contentTypeHeader), Value: []byte(s)})
	} else {
		b.Headers = append(b.Headers, sarama.RecordHeader{Key: []byte(prefix + attribute.Name()), Value: []byte(s)})
	}
	return nil
}

func (b *kafkaProducerMessageWriter) SetExtension(name string, value interface{}) error {
	if !b.skipKey && name == "key" {
		if v, ok := value.([]byte); ok {
			b.Key = sarama.ByteEncoder(v)
		} else {
			s, err := types.Format(value)
			if err != nil {
				return err
			}
			b.Key = sarama.ByteEncoder(s)
		}
		return nil
	}

	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.Headers = append(b.Headers, sarama.RecordHeader{Key: []byte(prefix + name), Value: []byte(s)})
	return nil
}

var _ binding.StructuredWriter = (*kafkaProducerMessageWriter)(nil) // Test it conforms to the interface
var _ binding.BinaryWriter = (*kafkaProducerMessageWriter)(nil)     // Test it conforms to the interface
