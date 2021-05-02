/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_segmentio_test

import (
	"context"
	"testing"

	"github.com/cloudevents/sdk-go/protocol/kafka_segmentio/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
)

// Avoid DCE
var M binding.Message
var Event *event.Event
var Err error

func BenchmarkNewStructuredMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = kafka_segmentio.NewMessageFromConsumerMessage(structuredConsumerMessage)
	}
}

func BenchmarkNewBinaryMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = kafka_segmentio.NewMessageFromConsumerMessage(binaryConsumerMessage)
	}
}

func BenchmarkNewStructuredMessageToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = kafka_segmentio.NewMessageFromConsumerMessage(structuredConsumerMessage)
		Event, Err = binding.ToEvent(context.TODO(), M)
	}
}

func BenchmarkNewBinaryMessageToEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		M = kafka_segmentio.NewMessageFromConsumerMessage(binaryConsumerMessage)
		Event, Err = binding.ToEvent(context.TODO(), M)
	}
}
