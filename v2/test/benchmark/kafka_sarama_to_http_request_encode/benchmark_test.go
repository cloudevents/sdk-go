package kafka_sarama_to_http_request_encode

import (
	"context"
	nethttp "net/http"
	"testing"

	test2 "github.com/cloudevents/sdk-go/v2/binding/test"

	"github.com/Shopify/sarama"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/cloudevents/sdk-go/v2/protocol/kafka_sarama"
)

var (
	event                     = test.FullEvent()
	structuredConsumerMessage = &sarama.ConsumerMessage{
		Value: test2.MustJSON(event),
		Headers: []*sarama.RecordHeader{{
			Key:   []byte("Content-Type"),
			Value: []byte(cloudevents.ApplicationCloudEventsJSON),
		}},
	}
	binaryConsumerMessage = &sarama.ConsumerMessage{
		Value: []byte("hello world!"),
		Headers: mustToSaramaConsumerHeaders(map[string]string{
			"ce_type":            event.Type(),
			"ce_source":          event.Source(),
			"ce_id":              event.ID(),
			"ce_time":            test.Timestamp.String(),
			"ce_specversion":     "1.0",
			"ce_dataschema":      test.Schema.String(),
			"ce_datacontenttype": "text/json",
			"ce_subject":         "topic",
			"ce_exta":            "someext",
		}),
	}
)

func mustToSaramaConsumerHeaders(m map[string]string) []*sarama.RecordHeader {
	res := make([]*sarama.RecordHeader, len(m))
	i := 0
	for k, v := range m {
		res[i] = &sarama.RecordHeader{Key: []byte(k), Value: []byte(v)}
		i++
	}
	return res
}

// Avoid DCE
var M binding.Message
var Req *nethttp.Request
var Err error

func BenchmarkStructured(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = kafka_sarama.NewMessageFromConsumerMessage(structuredConsumerMessage)
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		Err = http.WriteRequest(context.TODO(), M, Req, binding.TransformerFactories{})
	}
}

func BenchmarkBinary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = kafka_sarama.NewMessageFromConsumerMessage(binaryConsumerMessage)
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		Err = http.WriteRequest(context.TODO(), M, Req)
	}
}
