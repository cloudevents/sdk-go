/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http_request_to_kafka_sarama_encode

import (
	"context"
	nethttp "net/http"
	"testing"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

var (
	event                 cloudevents.Event
	structuredHttpRequest *nethttp.Request
	binaryHttpRequest     *nethttp.Request

	ctx = context.TODO()
)

func init() {
	event = test.FullEvent()

	structuredHttpRequest, _ = nethttp.NewRequest("POST", "http://localhost", nil)
	Err = http.WriteRequest(context.TODO(), (*binding.EventMessage)(&event), structuredHttpRequest)
	if Err != nil {
		panic(Err)
	}

	binaryHttpRequest, _ = nethttp.NewRequest("POST", "http://localhost", nil)
	Err = http.WriteRequest(context.TODO(), (*binding.EventMessage)(&event), binaryHttpRequest)
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
		Err = http.WriteRequest(context.TODO(), (*binding.EventMessage)(&event), Req)
	}
}

func BenchmarkStructured(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tempReq, _ := nethttp.NewRequest("POST", "http://localhost", nil)
		Err = http.WriteRequest(context.TODO(), (*binding.EventMessage)(&event), tempReq)

		M = http.NewMessageFromHttpRequest(tempReq)
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		producerMessage := &sarama.ProducerMessage{}
		Err = kafka_sarama.WriteProducerMessage(ctx, M, producerMessage)
	}
}

func BenchmarkBaselineBinary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Req, _ = nethttp.NewRequest("POST", "http://localhost", nil)
		Err = http.WriteRequest(context.TODO(), (*binding.EventMessage)(&event), Req)
	}
}

func BenchmarkBinary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tempReq, _ := nethttp.NewRequest("POST", "http://localhost", nil)
		Err = http.WriteRequest(context.TODO(), (*binding.EventMessage)(&event), tempReq)

		M = http.NewMessageFromHttpRequest(tempReq)
		Req, Err = nethttp.NewRequest("POST", "http://localhost", nil)
		producerMessage := &sarama.ProducerMessage{}
		Err = kafka_sarama.WriteProducerMessage(ctx, M, producerMessage)
	}
}
