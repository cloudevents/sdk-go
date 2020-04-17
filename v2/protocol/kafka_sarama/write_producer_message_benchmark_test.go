package kafka_sarama_test

import (
	"context"
	"testing"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol/kafka_sarama"
)

// Avoid DCE
var ProducerMessage *sarama.ProducerMessage

var (
	ctx               context.Context
	initialEvent      event.Event
	structuredMessage binding.Message
	binaryMessage     binding.Message
)

func init() {
	ctx = context.TODO()

	initialEvent = test.FullEvent()

	structuredMessage = test.MustCreateMockStructuredMessage(initialEvent)
	binaryMessage = test.MustCreateMockBinaryMessage(initialEvent)
}

func BenchmarkEncodeStructuredMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProducerMessage = &sarama.ProducerMessage{}
		Err = kafka_sarama.WriteProducerMessage(ctx, structuredMessage, ProducerMessage)
	}
}

func BenchmarkEncodeBinaryMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProducerMessage = &sarama.ProducerMessage{}
		Err = kafka_sarama.WriteProducerMessage(ctx, binaryMessage, ProducerMessage)
	}
}
