/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_sarama_test

import (
	"context"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
)

// A binary message carrying only the required attributes: no ce_time, no optional
// attributes and no extensions beyond the ones the binding synthesises itself.
func requiredOnlyConsumerMessage(specVersion string) *sarama.ConsumerMessage {
	return &sarama.ConsumerMessage{
		Topic: "test-topic",
		Value: []byte("hello world!"),
		Headers: mustToSaramaConsumerHeaders(map[string]string{
			"ce_specversion": specVersion,
			"ce_type":        testEvent.Type(),
			"ce_source":      testEvent.Source(),
			"ce_id":          testEvent.ID(),
		}),
	}
}

// binding.MessageMetadataReader documents that GetAttribute returns "attribute, nil"
// when the spec version supports the attribute kind but the message carries no value.
func TestGetAttributeReturnsNilWhenHeaderAbsent(t *testing.T) {
	m := kafka_sarama.NewMessageFromConsumerMessage(requiredOnlyConsumerMessage("1.0"))
	require.Equal(t, binding.EncodingBinary, m.ReadEncoding())

	attr, value := m.GetAttribute(spec.Time)
	require.NotNil(t, attr)
	require.Nil(t, value)
}

func TestGetExtensionReturnsNilWhenHeaderAbsent(t *testing.T) {
	m := kafka_sarama.NewMessageFromConsumerMessage(requiredOnlyConsumerMessage("1.0"))
	require.Nil(t, m.GetExtension("notthere"))
}

// transformer.AddTimeNow stamps the time attribute when the source message has none. It
// reaches the message through WriteProducerMessage, binding.Write and the binary transformer
// path, so a reader reporting "" instead of nil leaves its only branch unreachable.
func TestWriteProducerMessageAddTimeNow(t *testing.T) {
	var producerMessage sarama.ProducerMessage

	err := kafka_sarama.WriteProducerMessage(
		context.TODO(),
		kafka_sarama.NewMessageFromConsumerMessage(requiredOnlyConsumerMessage("1.0")),
		&producerMessage,
		transformer.AddTimeNow,
	)
	require.NoError(t, err)

	found := false
	for _, h := range producerMessage.Headers {
		if string(h.Key) == "ce_time" {
			found = true
			require.NotEmpty(t, string(h.Value))
		}
	}
	require.True(t, found, "WriteProducerMessage did not add the ce_time header")
}

// transformer.Version rewrites an attribute only when GetAttribute reports a value, so a
// reader reporting "" for every absent attribute makes a version conversion emit one empty
// header per attribute the event does not carry. For subject, CloudEvents v1.0.2 says
// "If present, MUST be a non-empty string", and for dataschema "If present, MUST be a
// non-empty URI", so those headers must be left out rather than written empty.
func TestWriteProducerMessageVersionDoesNotEmitEmptyAttributes(t *testing.T) {
	var producerMessage sarama.ProducerMessage

	err := kafka_sarama.WriteProducerMessage(
		context.TODO(),
		kafka_sarama.NewMessageFromConsumerMessage(requiredOnlyConsumerMessage("0.3")),
		&producerMessage,
		transformer.Version(spec.V1),
	)
	require.NoError(t, err)

	var empty []string
	for _, h := range producerMessage.Headers {
		if len(h.Value) == 0 {
			empty = append(empty, string(h.Key))
		}
	}
	require.Empty(t, empty, "headers written with an empty value")
}
