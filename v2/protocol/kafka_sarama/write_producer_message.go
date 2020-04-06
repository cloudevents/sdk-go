package kafka_sarama

import (
	"bytes"
	"context"
	"io"

	"github.com/Shopify/sarama"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/types"
)

const (
	partitionKey = "partitionkey"
)

// WriteProducerMessage fills the provided producerMessage with the message m.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
func WriteProducerMessage(ctx context.Context, m binding.Message, producerMessage *sarama.ProducerMessage, transformers ...binding.Transformer) error {
	enc := (*kafkaProducerMessageWriter)(producerMessage)
// By default, this function implements the key mapping, trying to set the key of the message based on partitionKey extension.
// If you want to disable the Key Mapping, decorate the context with `WithSkipKeyMapping`
func WriteProducerMessage(ctx context.Context, m binding.Message, producerMessage *sarama.ProducerMessage, transformers ...binding.TransformerFactory) error {
	writer := (*kafkaProducerMessageWriter)(producerMessage)

	skipKey := binding.GetOrDefaultFromCtx(ctx, skipKeyKey{}, false).(bool)

	// If skipKey = true, then we can just use the default write algorithm
	if skipKey {
		_, err := binding.Write(
			ctx,
			m,
			writer,
			writer,
			transformers...,
		)
		return err
	}

	// if skipKey = false, we can't accept structured encoding. We can
	encoding := m.ReadEncoding()
	var err error
	// Try direct encoding only if the event is a binary event
	if encoding == binding.EncodingBinary {
		// Specialized writer that writes the key
		writer := (*kafkaProducerMessageWithKeyWriter)(producerMessage)
		encoding, err = binding.DirectWrite(ctx, m, nil, writer, transformers...)
		if encoding != binding.EncodingUnknown {
			// Message directly encoded binary -> binary, nothing else to do here
			return err
		}
	}

	var e *ce.Event
	e, err = binding.ToEvent(ctx, m, transformers...)
	if err != nil {
		return err
	}

	// Let's check if event contains a partitionKey to write
	if val, ok := e.Extensions()[partitionKey]; ok {
		s, err := types.Format(val)
		if err != nil {
			return err
		}
		producerMessage.Key = sarama.StringEncoder(s)
	}

	_, err = binding.WriteEvent(ctx, e, writer, writer)
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
	if attribute.Kind() == spec.DataContentType {
		if value == nil {
			b.removeHeader(contentTypeHeader)
			return nil
		}

		// Everything is a string here
		s, err := types.Format(value)
		if err != nil {
			return err
		}
		b.Headers = append(b.Headers, sarama.RecordHeader{Key: []byte(contentTypeHeader), Value: []byte(s)})
	} else {
		if value == nil {
			b.removeHeader(prefix + attribute.Name())
			return nil
		}

		// Everything is a string here
		s, err := types.Format(value)
		if err != nil {
			return err
		}
		b.Headers = append(b.Headers, sarama.RecordHeader{Key: []byte(prefix + attribute.Name()), Value: []byte(s)})
	}
	return nil
}

func (b *kafkaProducerMessageWriter) SetExtension(name string, value interface{}) error {
	if value == nil {
		b.removeHeader(prefix + name)
		return nil
	}

	// Kafka headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.Headers = append(b.Headers, sarama.RecordHeader{Key: []byte(prefix + name), Value: []byte(s)})
	return nil
}

func (b *kafkaProducerMessageWriter) removeHeader(name string) {
	k := []byte(name)
	for index, h := range b.Headers {
		if bytes.Equal(k, h.Key) {
			b.Headers = append(b.Headers[:index], b.Headers[index+1:]...)
			return
		}
	}
}

var _ binding.StructuredWriter = (*kafkaProducerMessageWriter)(nil) // Test it conforms to the interface
var _ binding.BinaryWriter = (*kafkaProducerMessageWriter)(nil)     // Test it conforms to the interface
