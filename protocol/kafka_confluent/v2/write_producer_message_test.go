/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package kafka_confluent

import (
	"context"
	"strconv"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	. "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	. "github.com/cloudevents/sdk-go/v2/test"
)

func TestWriteProducerMessage(t *testing.T) {
	tests := []struct {
		name             string
		context          context.Context
		messageFactory   func(e event.Event) binding.Message
		expectedEncoding binding.Encoding
	}{
		{
			name:    "Structured to Structured",
			context: ctx,
			messageFactory: func(e event.Event) binding.Message {
				return MustCreateMockStructuredMessage(t, e)
			},
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Binary to Binary",
			context:          ctx,
			messageFactory:   MustCreateMockBinaryMessage,
			expectedEncoding: binding.EncodingBinary,
		},
	}
	EachEvent(t, Events(), func(t *testing.T, e event.Event) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx := tt.context
				topic := "test-topic"
				kafkaMessage := &kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     &topic,
						Partition: int32(0),
						Offset:    kafka.Offset(10),
					},
				}

				eventIn := ConvertEventExtensionsToString(t, e.Clone())
				messageIn := tt.messageFactory(eventIn)

				err := WriteProducerMessage(ctx, messageIn, kafkaMessage)
				require.NoError(t, err)

				messageOut := NewMessage(kafkaMessage)
				require.Equal(t, tt.expectedEncoding, messageOut.ReadEncoding())

				if tt.expectedEncoding == binding.EncodingBinary {
					err = messageOut.ReadBinary(ctx, (*kafkaMessageWriter)(kafkaMessage))
				}
				require.NoError(t, err)

				eventOut, err := binding.ToEvent(ctx, messageOut)
				require.NoError(t, err)
				if tt.expectedEncoding == binding.EncodingBinary {
					eventIn.SetExtension(KafkaPartitionKey, strconv.FormatInt(int64(kafkaMessage.TopicPartition.Partition), 10))
					eventIn.SetExtension(KafkaOffsetKey, strconv.FormatInt(int64(kafkaMessage.TopicPartition.Offset), 10))
					eventIn.SetExtension(KafkaTopicKey, kafkaMessage.TopicPartition.Topic)
				}
				AssertEventEquals(t, eventIn, *eventOut)
			})
		}
	})
}
