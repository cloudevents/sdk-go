/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_franz

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"

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
			name:    "structured to structured",
			context: ctx,
			messageFactory: func(e event.Event) binding.Message {
				return MustCreateMockStructuredMessage(t, e)
			},
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "binary to binary",
			context:          ctx,
			messageFactory:   MustCreateMockBinaryMessage,
			expectedEncoding: binding.EncodingBinary,
		},
	}

	EachEvent(t, Events(), func(t *testing.T, e event.Event) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				record := &kgo.Record{
					Topic:     "test-topic",
					Partition: 0,
					Offset:    10,
				}

				eventIn := ConvertEventExtensionsToString(t, e.Clone())
				messageIn := tt.messageFactory(eventIn)

				err := WriteProducerMessage(tt.context, messageIn, record)
				require.NoError(t, err)

				messageOut := NewMessage(record)
				require.Equal(t, tt.expectedEncoding, messageOut.ReadEncoding())

				if tt.expectedEncoding == binding.EncodingBinary {
					err = messageOut.ReadBinary(tt.context, (*kafkaRecordWriter)(record))
					require.NoError(t, err)
				}

				eventOut, err := binding.ToEvent(tt.context, messageOut)
				require.NoError(t, err)

				if tt.expectedEncoding == binding.EncodingBinary {
					eventIn.SetExtension(KafkaPartitionKey, strconv.FormatInt(int64(record.Partition), 10))
					eventIn.SetExtension(KafkaOffsetKey, strconv.FormatInt(record.Offset, 10))
					eventIn.SetExtension(KafkaTopicKey, record.Topic)
				}
				AssertEventEquals(t, eventIn, *eventOut)
			})
		}
	})
}
