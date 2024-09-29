/*
 Copyright 2023 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_confluent

import (
	"context"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/require"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/test"
)

var (
	ctx            = context.Background()
	testEvent      = test.FullEvent()
	testTopic      = "test-topic"
	topicPartition = kafka.TopicPartition{
		Topic:     &testTopic,
		Partition: int32(0),
		Offset:    kafka.Offset(10),
	}
	structuredConsumerMessage = &kafka.Message{
		TopicPartition: topicPartition,
		Value: func() []byte {
			b, _ := format.JSON.Marshal(&testEvent)
			return b
		}(),
		Headers: []kafka.Header{{
			Key:   "content-type",
			Value: []byte(cloudevents.ApplicationCloudEventsJSON),
		}},
	}
	binaryConsumerMessage = &kafka.Message{
		TopicPartition: topicPartition,
		Value:          []byte("hello world!"),
		Headers: mapToKafkaHeaders(map[string]string{
			"ce_type":            testEvent.Type(),
			"ce_source":          testEvent.Source(),
			"ce_id":              testEvent.ID(),
			"ce_time":            test.Timestamp.String(),
			"ce_specversion":     "1.0",
			"ce_dataschema":      test.Schema.String(),
			"ce_datacontenttype": "text/json",
			"ce_subject":         "receiverTopic",
			"exta":               "someext",
		}),
	}
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name             string
		consumerMessage  *kafka.Message
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
			consumerMessage: &kafka.Message{
				TopicPartition: topicPartition,
				Value:          []byte("{}"),
				Headers: []kafka.Header{{
					Key:   "content-type",
					Value: []byte("application/json"),
				}},
			},
			expectedEncoding: binding.EncodingUnknown,
		},
		{
			name: "Binary encoding with empty value",
			consumerMessage: &kafka.Message{
				TopicPartition: topicPartition,
				Value:          nil,
				Headers: mapToKafkaHeaders(map[string]string{
					"ce_type":            testEvent.Type(),
					"ce_source":          testEvent.Source(),
					"ce_id":              testEvent.ID(),
					"ce_time":            test.Timestamp.String(),
					"ce_specversion":     "1.0",
					"ce_dataschema":      test.Schema.String(),
					"ce_datacontenttype": "text/json",
					"ce_subject":         "receiverTopic",
				}),
			},
			expectedEncoding: binding.EncodingBinary,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewMessage(tt.consumerMessage)
			require.Equal(t, tt.expectedEncoding, msg.ReadEncoding())

			var err error
			if tt.expectedEncoding == binding.EncodingStructured {
				err = msg.ReadStructured(ctx, (*kafkaMessageWriter)(tt.consumerMessage))
			}

			if tt.expectedEncoding == binding.EncodingBinary {
				err = msg.ReadBinary(ctx, (*kafkaMessageWriter)(tt.consumerMessage))
			}
			require.Nil(t, err)
		})
	}
}

func mapToKafkaHeaders(m map[string]string) []kafka.Header {
	res := make([]kafka.Header, len(m))
	i := 0
	for k, v := range m {
		res[i] = kafka.Header{Key: k, Value: []byte(v)}
		i++
	}
	return res
}
