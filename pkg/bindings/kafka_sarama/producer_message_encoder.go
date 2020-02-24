package kafka_sarama

import (
	"bytes"
	"context"
	"io"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

const (
	SKIP_KEY_EXTENSION = "SKIP_KEY_EXTENSION"
)

// Fill the provided producerMessage with the message m.
// Using context you can tweak the encoding processing (more details on binding.Translate documentation).
func EncodeKafkaProducerMessage(ctx context.Context, m binding.Message, producerMessage *sarama.ProducerMessage, transformerFactories binding.TransformerFactories) error {
	skipKey := binding.GetOrDefaultFromCtx(ctx, SKIP_KEY_EXTENSION, false).(bool)

	enc := &kafkaProducerMessageEncoder{
		producerMessage,
		skipKey,
	}

	if skipKey {
		_, err := binding.Encode(
			ctx,
			m,
			enc,
			enc,
			transformerFactories,
		)
		return err
	}

	_, err := binding.Encode(
		ctx,
		m,
		nil,
		enc,
		transformerFactories,
	)
	return err
}

type kafkaProducerMessageEncoder struct {
	*sarama.ProducerMessage
	skipKey bool
}

func (b *kafkaProducerMessageEncoder) SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) error {
	b.Headers = []sarama.RecordHeader{{
		Key:   []byte(ContentType),
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

func (b *kafkaProducerMessageEncoder) Start(ctx context.Context) error {
	b.Headers = []sarama.RecordHeader{}
	return nil
}

func (b *kafkaProducerMessageEncoder) End() error {
	return nil
}

func (b *kafkaProducerMessageEncoder) SetData(reader io.Reader) error {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, reader)
	if err != nil {
		return err
	}

	b.Value = sarama.ByteEncoder(buf.Bytes())
	return nil
}

func (b *kafkaProducerMessageEncoder) SetAttribute(attribute spec.Attribute, value interface{}) error {
	// Everything is a string here
	s, err := types.Format(value)
	if err != nil {
		return err
	}

	if attribute.Kind() == spec.DataContentType {
		b.Headers = append(b.Headers, sarama.RecordHeader{Key: []byte(ContentType), Value: []byte(s)})
	} else {
		b.Headers = append(b.Headers, sarama.RecordHeader{Key: []byte(prefix + attribute.Name()), Value: []byte(s)})
	}
	return nil
}

func (b *kafkaProducerMessageEncoder) SetExtension(name string, value interface{}) error {
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

var _ binding.StructuredEncoder = (*kafkaProducerMessageEncoder)(nil) // Test it conforms to the interface
var _ binding.BinaryEncoder = (*kafkaProducerMessageEncoder)(nil)     // Test it conforms to the interface

func WithSkipKeyExtension(ctx context.Context) context.Context {
	return context.WithValue(ctx, SKIP_KEY_EXTENSION, true)
}
