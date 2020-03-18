// +build kafka

package kafka_sarama_test

import (
	"context"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v1"
	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/bindings/kafka_sarama"
)

// Avoid DCE
var M binding.Message
var Event cloudevents.Event
var Err error

func BenchmarkNewStructuredMessageWithoutKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(kafka_sarama.StructuredConsumerMessageWithoutKey)
	}
}

func BenchmarkNewStructuredMessageWithKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(kafka_sarama.StructuredConsumerMessageWithKey)
	}
}

func BenchmarkNewBinaryMessageWithoutKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(kafka_sarama.BinaryConsumerMessageWithoutKey)
	}
}

func BenchmarkNewBinaryMessageWithKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(kafka_sarama.BinaryConsumerMessageWithKey)
	}
}

func BenchmarkNewStructuredMessageWithoutKeyToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(kafka_sarama.StructuredConsumerMessageWithoutKey)
		Event, _, Err = binding.ToEvent(context.TODO(), M)
	}
}

func BenchmarkNewStructuredMessageWithKeyToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(kafka_sarama.StructuredConsumerMessageWithKey)
		Event, _, Err = binding.ToEvent(context.TODO(), M)
	}
}

func BenchmarkNewBinaryMessageWithoutKeyToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(kafka_sarama.BinaryConsumerMessageWithoutKey)
		Event, _, Err = binding.ToEvent(context.TODO(), M)
	}
}

func BenchmarkNewBinaryMessageWithKeyToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(kafka_sarama.BinaryConsumerMessageWithKey)
		Event, _, Err = binding.ToEvent(context.TODO(), M)
	}
}
