// +build kafka

package kafka_sarama_test

import (
	"context"
	"testing"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/bindings/kafka_sarama"
	"github.com/cloudevents/sdk-go/pkg/event"
)

// Avoid DCE
var ProducerMessage *sarama.ProducerMessage

var (
	ctxSkipKey                  context.Context
	ctx                         context.Context
	eventWithoutKey             event.Event
	eventWithKey                event.Event
	structuredMessageWithoutKey binding.Message
	structuredMessageWithKey    binding.Message
	binaryMessageWithoutKey     binding.Message
	binaryMessageWithKey        binding.Message
)

func init() {
	ctxSkipKey = kafka_sarama.WithSkipKeyExtension(context.TODO())
	ctx = context.TODO()

	eventWithoutKey = test.FullEvent()
	eventWithKey = test.FullEvent()
	eventWithKey.SetExtension("key", "aaaaaa")

	structuredMessageWithoutKey = test.MustCreateMockStructuredMessage(eventWithoutKey)
	structuredMessageWithKey = test.MustCreateMockStructuredMessage(eventWithKey)
	binaryMessageWithoutKey = test.MustCreateMockBinaryMessage(eventWithoutKey)
	binaryMessageWithKey = test.MustCreateMockBinaryMessage(eventWithKey)
}

func BenchmarkEncodeStructuredMessageSkipKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProducerMessage = &sarama.ProducerMessage{}
		Err = kafka_sarama.EncodeKafkaProducerMessage(ctxSkipKey, structuredMessageWithoutKey, ProducerMessage, binding.TransformerFactories{})
	}
}

func BenchmarkEncodeStructuredMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProducerMessage = &sarama.ProducerMessage{}
		Err = kafka_sarama.EncodeKafkaProducerMessage(ctx, structuredMessageWithKey, ProducerMessage, binding.TransformerFactories{})
	}
}

func BenchmarkEncodeBinaryMessageSkipKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProducerMessage = &sarama.ProducerMessage{}
		Err = kafka_sarama.EncodeKafkaProducerMessage(ctxSkipKey, binaryMessageWithoutKey, ProducerMessage, binding.TransformerFactories{})
	}
}

func BenchmarkEncodeBinaryMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProducerMessage = &sarama.ProducerMessage{}
		Err = kafka_sarama.EncodeKafkaProducerMessage(ctx, binaryMessageWithKey, ProducerMessage, binding.TransformerFactories{})
	}
}
