package kafka_sarama_test

import (
	"context"
	"testing"

	"github.com/Shopify/sarama"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol/kafka_sarama"
)

// Avoid DCE
var M binding.Message
var Event *event.Event
var Err error

var (
	benchEvent                = test.FullEvent()
	structuredConsumerMessage = &sarama.ConsumerMessage{
		Value: test.MustJSON(benchEvent),
		Headers: []*sarama.RecordHeader{{
			Key:   []byte("content-type"),
			Value: []byte(cloudevents.ApplicationCloudEventsJSON),
		}},
	}
	binaryConsumerMessage = &sarama.ConsumerMessage{
		Value: []byte("hello world!"),
		Headers: mustToSaramaConsumerHeaders(map[string]string{
			"ce_type":            benchEvent.Type(),
			"ce_source":          benchEvent.Source(),
			"ce_id":              benchEvent.ID(),
			"ce_time":            test.Timestamp.String(),
			"ce_specversion":     "1.0",
			"ce_dataschema":      test.Schema.String(),
			"ce_datacontenttype": "text/json",
			"ce_subject":         "receiverTopic",
			"ce_exta":            "someext",
		}),
	}
)

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

func mustToSaramaConsumerHeaders(m map[string]string) []*sarama.RecordHeader {
	res := make([]*sarama.RecordHeader, len(m))
	i := 0
	for k, v := range m {
		res[i] = &sarama.RecordHeader{Key: []byte(k), Value: []byte(v)}
		i++
	}
	return res
}
