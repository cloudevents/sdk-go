package kafka_sarama_test

import (
	"context"
	"testing"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/transport/kafka_sarama"
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
		Err = kafka_sarama.WriteKafkaProducerMessage(ctxSkipKey, structuredMessageWithoutKey, ProducerMessage, nil)
	}
}

func BenchmarkEncodeStructuredMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProducerMessage = &sarama.ProducerMessage{}
		Err = kafka_sarama.WriteKafkaProducerMessage(ctx, structuredMessageWithKey, ProducerMessage, nil)
	}
}

func BenchmarkEncodeBinaryMessageSkipKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProducerMessage = &sarama.ProducerMessage{}
		Err = kafka_sarama.WriteKafkaProducerMessage(ctxSkipKey, binaryMessageWithoutKey, ProducerMessage, nil)
	}
}

func BenchmarkEncodeBinaryMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProducerMessage = &sarama.ProducerMessage{}
		Err = kafka_sarama.WriteKafkaProducerMessage(ctx, binaryMessageWithKey, ProducerMessage, nil)
	}
}
