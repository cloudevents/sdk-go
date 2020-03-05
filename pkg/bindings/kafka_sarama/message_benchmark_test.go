package kafka_sarama_test

import (
	"context"
	"testing"

	"github.com/Shopify/sarama"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/bindings/kafka_sarama"
	"github.com/cloudevents/sdk-go/pkg/event"
)

// Avoid DCE
var M binding.Message
var Event event.Event
var Err error

var (
	e                                   = test.FullEvent()
	structuredConsumerMessageWithoutKey = &sarama.ConsumerMessage{
		Value: test.MustJSON(e),
		Headers: []*sarama.RecordHeader{{
			Key:   []byte(kafka_sarama.ContentType),
			Value: []byte(cloudevents.ApplicationCloudEventsJSON),
		}},
	}
	structuredConsumerMessageWithKey = &sarama.ConsumerMessage{
		Key:   []byte("aaa"),
		Value: test.MustJSON(e),
		Headers: []*sarama.RecordHeader{{
			Key:   []byte(kafka_sarama.ContentType),
			Value: []byte(cloudevents.ApplicationCloudEventsJSON),
		}},
	}
	binaryConsumerMessageWithoutKey = &sarama.ConsumerMessage{
		Value: []byte("hello world!"),
		Headers: mustToSaramaConsumerHeaders(map[string]string{
			"ce_type":            e.Type(),
			"ce_source":          e.Source(),
			"ce_id":              e.ID(),
			"ce_time":            test.Timestamp.String(),
			"ce_specversion":     "1.0",
			"ce_dataschema":      test.Schema.String(),
			"ce_datacontenttype": "text/json",
			"ce_subject":         "topic",
			"ce_exta":            "someext",
		}),
	}
	binaryConsumerMessageWithKey = &sarama.ConsumerMessage{
		Key:   []byte("akey"),
		Value: []byte("hello world!"),
		Headers: mustToSaramaConsumerHeaders(map[string]string{
			"ce_type":            e.Type(),
			"ce_source":          e.Source(),
			"ce_id":              e.ID(),
			"ce_time":            test.Timestamp.String(),
			"ce_specversion":     "1.0",
			"ce_dataschema":      test.Schema.String(),
			"ce_datacontenttype": "text/json",
			"ce_subject":         "topic",
			"ce_exta":            "someext",
		}),
	}
)

func BenchmarkNewStructuredMessageWithoutKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(structuredConsumerMessageWithoutKey)
	}
}

func BenchmarkNewStructuredMessageWithKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(structuredConsumerMessageWithKey)
	}
}

func BenchmarkNewBinaryMessageWithoutKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(binaryConsumerMessageWithoutKey)
	}
}

func BenchmarkNewBinaryMessageWithKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(binaryConsumerMessageWithKey)
	}
}

func BenchmarkNewStructuredMessageWithoutKeyToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(structuredConsumerMessageWithoutKey)
		Event, _, Err = binding.ToEvent(context.TODO(), M)
	}
}

func BenchmarkNewStructuredMessageWithKeyToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(structuredConsumerMessageWithKey)
		Event, _, Err = binding.ToEvent(context.TODO(), M)
	}
}

func BenchmarkNewBinaryMessageWithoutKeyToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(binaryConsumerMessageWithoutKey)
		Event, _, Err = binding.ToEvent(context.TODO(), M)
	}
}

func BenchmarkNewBinaryMessageWithKeyToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M, Err = kafka_sarama.NewMessage(binaryConsumerMessageWithKey)
		Event, _, Err = binding.ToEvent(context.TODO(), M)
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
