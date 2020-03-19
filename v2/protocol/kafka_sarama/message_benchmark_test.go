package kafka_sarama_test

import (
	"context"
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol/kafka_sarama"
)

// Avoid DCE
var M binding.Message
var Event *event.Event
var Err error

func BenchmarkNewStructuredMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = kafka_sarama.NewMessageFromConsumerMessage(structuredConsumerMessage)
	}
}

func BenchmarkNewBinaryMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = kafka_sarama.NewMessageFromConsumerMessage(binaryConsumerMessage)
	}
}

func BenchmarkNewStructuredMessageToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = kafka_sarama.NewMessageFromConsumerMessage(structuredConsumerMessage)
		Event, Err = binding.ToEvent(context.TODO(), M)
	}
}

func BenchmarkNewBinaryMessageToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = kafka_sarama.NewMessageFromConsumerMessage(binaryConsumerMessage)
		Event, Err = binding.ToEvent(context.TODO(), M)
	}
}
