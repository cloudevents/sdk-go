package kafka_sarama

import (
	"bytes"
	"context"
	"io"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/types"
)

// Fill the provided producerMessage with the message m.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
func WriteProducerMessage(ctx context.Context, m binding.Message, producerMessage *sarama.ProducerMessage, transformerFactories ...binding.TransformerFactory) error {
	enc := (*kafkaProducerMessageWriter)(producerMessage)

	_, err := binding.Write(
		ctx,
		m,
		enc,
		enc,
		transformerFactories...,
	)
	return err
}

type kafkaProducerMessageWriter sarama.ProducerMessage

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

func (b *kafkaProducerMessageWriter) End(ctx context.Context) error {
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
	// Kafka headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.Headers = append(b.Headers, sarama.RecordHeader{Key: []byte(prefix + name), Value: []byte(s)})
	return nil
}

var _ binding.StructuredWriter = (*kafkaProducerMessageWriter)(nil) // Test it conforms to the interface
var _ binding.BinaryWriter = (*kafkaProducerMessageWriter)(nil)     // Test it conforms to the interface
