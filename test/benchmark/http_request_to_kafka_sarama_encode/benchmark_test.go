package kafka_sarama_to_http_request_encode

import (
	"context"
	nethttp "net/http"
	"testing"

	"github.com/Shopify/sarama"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/transport/kafka_sarama"
	"github.com/cloudevents/sdk-go/pkg/transport/test"
)

var (
	eventWithoutKey                 cloudevents.Event
	eventWithKey                    cloudevents.Event
	structuredHttpRequestWithoutKey *nethttp.Request
	structuredHttpRequestWithKey    *nethttp.Request
	binaryHttpRequestWithoutKey     *nethttp.Request
	binaryHttpRequestWithKey        *nethttp.Request

	ctxSkipKey = kafka_sarama.WithSkipKeyExtension(context.TODO())
	ctx        = context.TODO()
)

func init() {
	eventWithoutKey = test.FullEvent()
	eventWithKey = test.FullEvent()
	eventWithKey.SetExtension("key", "aaa")

	structuredHttpRequestWithoutKey, _ = nethttp.NewRequest("POST", "http://localhost", nil)
	Err = http.WriteHttpRequest(context.TODO(), binding.EventMessage(eventWithoutKey), structuredHttpRequestWithoutKey, binding.TransformerFactories{})
	if Err != nil {
		panic(Err)
	}

	structuredHttpRequestWithKey, _ = nethttp.NewRequest("POST", "http://localhost", nil)
	Err = http.WriteHttpRequest(context.TODO(), binding.EventMessage(eventWithKey), structuredHttpRequestWithKey, binding.TransformerFactories{})
	if Err != nil {
		panic(Err)
	}

	binaryHttpRequestWithoutKey, _ = nethttp.NewRequest("POST", "http://localhost", nil)
	Err = http.WriteHttpRequest(context.TODO(), binding.EventMessage(eventWithoutKey), binaryHttpRequestWithoutKey, binding.TransformerFactories{})
	if Err != nil {
		panic(Err)
	}

	binaryHttpRequestWithKey, _ = nethttp.NewRequest("POST", "http://localhost", nil)
	Err = http.WriteHttpRequest(context.TODO(), binding.EventMessage(eventWithKey), binaryHttpRequestWithKey, binding.TransformerFactories{})
	if Err != nil {
		panic(Err)
	}

}

// Avoid DCE
var M binding.Message
var Req *nethttp.Request
var Err error

func BenchmarkBaselineStructured(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Req, _ = nethttp.NewRequest("POST", "http://localhost", nil)
		Err = http.WriteHttpRequest(context.TODO(), binding.EventMessage(eventWithKey), Req, binding.TransformerFactories{})
	}
}

func BenchmarkStructured(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tempReq, _ := nethttp.NewRequest("POST", "http://localhost", nil)
		Err = http.WriteHttpRequest(context.TODO(), binding.EventMessage(eventWithKey), tempReq, binding.TransformerFactories{})

		M = http.NewMessageFromHttpRequest(tempReq)
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		producerMessage := &sarama.ProducerMessage{}
		Err = kafka_sarama.WriteKafkaProducerMessage(ctx, M, producerMessage, binding.TransformerFactories{})
	}
}

func BenchmarkBaselineStructuredSkipKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Req, _ = nethttp.NewRequest("POST", "http://localhost", nil)
		Err = http.WriteHttpRequest(context.TODO(), binding.EventMessage(eventWithoutKey), Req, binding.TransformerFactories{})
	}
}

func BenchmarkStructuredSkipKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tempReq, _ := nethttp.NewRequest("POST", "http://localhost", nil)
		Err = http.WriteHttpRequest(context.TODO(), binding.EventMessage(eventWithoutKey), tempReq, binding.TransformerFactories{})

		M = http.NewMessageFromHttpRequest(tempReq)
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		producerMessage := &sarama.ProducerMessage{}
		Err = kafka_sarama.WriteKafkaProducerMessage(ctxSkipKey, M, producerMessage, binding.TransformerFactories{})
	}
}

func BenchmarkBaselineBinary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Req, _ = nethttp.NewRequest("POST", "http://localhost", nil)
		Err = http.WriteHttpRequest(context.TODO(), binding.EventMessage(eventWithKey), Req, binding.TransformerFactories{})
	}
}

func BenchmarkBinary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tempReq, _ := nethttp.NewRequest("POST", "http://localhost", nil)
		Err = http.WriteHttpRequest(context.TODO(), binding.EventMessage(eventWithKey), tempReq, binding.TransformerFactories{})

		M = http.NewMessageFromHttpRequest(tempReq)
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		producerMessage := &sarama.ProducerMessage{}
		Err = kafka_sarama.WriteKafkaProducerMessage(ctx, M, producerMessage, binding.TransformerFactories{})
	}
}

func BenchmarkBaselineBinarySkipKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Req, _ = nethttp.NewRequest("POST", "http://localhost", nil)
		Err = http.WriteHttpRequest(context.TODO(), binding.EventMessage(eventWithoutKey), Req, binding.TransformerFactories{})
	}
}

func BenchmarkBinarySkipKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tempReq, _ := nethttp.NewRequest("POST", "http://localhost", nil)
		Err = http.WriteHttpRequest(context.TODO(), binding.EventMessage(eventWithoutKey), tempReq, binding.TransformerFactories{})

		M = http.NewMessageFromHttpRequest(tempReq)
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		producerMessage := &sarama.ProducerMessage{}
		Err = kafka_sarama.WriteKafkaProducerMessage(ctxSkipKey, M, producerMessage, binding.TransformerFactories{})
	}
}
