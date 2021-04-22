/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_sarama_test

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/test"
)

var (
	testEvent                 = test.FullEvent()
	structuredConsumerMessage = &sarama.ConsumerMessage{
		Value: func() []byte {
			b, _ := format.JSON.Marshal(&testEvent)
			return b
		}(),
		Headers: []*sarama.RecordHeader{{
			Key:   []byte("content-type"),
			Value: []byte(cloudevents.ApplicationCloudEventsJSON),
		}},
	}
	binaryConsumerMessage = &sarama.ConsumerMessage{
		Value: []byte("hello world!"),
		Headers: mustToSaramaConsumerHeaders(map[string]string{
			"ce_type":            testEvent.Type(),
			"ce_source":          testEvent.Source(),
			"ce_id":              testEvent.ID(),
			"ce_time":            test.Timestamp.String(),
			"ce_specversion":     "1.0",
			"ce_dataschema":      test.Schema.String(),
			"ce_datacontenttype": "text/json",
			"ce_subject":         "receiverTopic",
			"ce_exta":            "someext",
		}),
	}
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name             string
		consumerMessage  *sarama.ConsumerMessage
		expectedEncoding binding.Encoding
	}{
		{
			name:             "Structured encoding",
			consumerMessage:  structuredConsumerMessage,
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Binary encoding",
			consumerMessage:  binaryConsumerMessage,
			expectedEncoding: binding.EncodingBinary,
		},
		{
			name: "Unknown encoding",
			consumerMessage: &sarama.ConsumerMessage{
				Value: []byte("{}"),
				Headers: []*sarama.RecordHeader{{
					Key:   []byte("content-type"),
					Value: []byte("application/json"),
				}},
			},
			expectedEncoding: binding.EncodingUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := kafka_sarama.NewMessageFromConsumerMessage(tt.consumerMessage)
			require.NotNil(t, got)
			require.Equal(t, tt.expectedEncoding, got.ReadEncoding())
		})
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
