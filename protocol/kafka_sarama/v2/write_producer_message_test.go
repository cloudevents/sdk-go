/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_sarama

import (
	"context"
	"strings"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	. "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	. "github.com/cloudevents/sdk-go/v2/test"
)

const testKey = "hello-key"

func TestEncodeKafkaProducerMessage(t *testing.T) {
	tests := []struct {
		name             string
		context          context.Context
		addPartitionKey  bool
		messageFactory   func(e event.Event) binding.Message
		expectedEncoding binding.Encoding
		expectedKey      bool
	}{
		{
			name:    "Structured to Structured - skip key mapping",
			context: WithSkipKeyMapping(context.TODO()),
			messageFactory: func(e event.Event) binding.Message {
				return MustCreateMockStructuredMessage(t, e)
			},
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Binary to Binary - skip key mapping",
			context:          WithSkipKeyMapping(context.TODO()),
			messageFactory:   MustCreateMockBinaryMessage,
			expectedEncoding: binding.EncodingBinary,
		},
		{
			name:             "Event to Structured - skip key mapping",
			context:          WithSkipKeyMapping(binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured)),
			messageFactory:   func(e event.Event) binding.Message { return (*binding.EventMessage)(&e) },
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Event to Binary - skip key mapping",
			context:          WithSkipKeyMapping(binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary)),
			messageFactory:   func(e event.Event) binding.Message { return (*binding.EventMessage)(&e) },
			expectedEncoding: binding.EncodingBinary,
		},
		{
			name:            "Structured to Structured - with key & skip key mapping",
			context:         WithSkipKeyMapping(context.TODO()),
			addPartitionKey: true,
			messageFactory: func(e event.Event) binding.Message {
				return MustCreateMockStructuredMessage(t, e)
			},
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Binary to Binary - with key & skip key mapping",
			context:          WithSkipKeyMapping(context.TODO()),
			addPartitionKey:  true,
			messageFactory:   MustCreateMockBinaryMessage,
			expectedEncoding: binding.EncodingBinary,
		},
		{
			name:             "Event to Structured - with key & skip key mapping",
			context:          WithSkipKeyMapping(binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured)),
			addPartitionKey:  true,
			messageFactory:   func(e event.Event) binding.Message { return (*binding.EventMessage)(&e) },
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Event to Binary - with key & skip key mapping",
			context:          WithSkipKeyMapping(binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary)),
			addPartitionKey:  true,
			messageFactory:   func(e event.Event) binding.Message { return (*binding.EventMessage)(&e) },
			expectedEncoding: binding.EncodingBinary,
		},
		{
			name:    "Structured to Structured - no key",
			context: binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			messageFactory: func(e event.Event) binding.Message {
				return MustCreateMockStructuredMessage(t, e)
			},
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Binary to Binary - no key",
			context:          context.TODO(),
			messageFactory:   MustCreateMockBinaryMessage,
			expectedEncoding: binding.EncodingBinary,
		},
		{
			name:             "Event to Structured - no key",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			messageFactory:   func(e event.Event) binding.Message { return (*binding.EventMessage)(&e) },
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Event to Binary - no key",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary),
			messageFactory:   func(e event.Event) binding.Message { return (*binding.EventMessage)(&e) },
			expectedEncoding: binding.EncodingBinary,
		},
		{
			name:            "Structured to Structured - with key",
			context:         binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			addPartitionKey: true,
			messageFactory: func(e event.Event) binding.Message {
				return MustCreateMockStructuredMessage(t, e)
			},
			expectedEncoding: binding.EncodingStructured,
			expectedKey:      true,
		},
		{
			name:             "Binary to Binary - with key",
			context:          context.TODO(),
			addPartitionKey:  true,
			messageFactory:   MustCreateMockBinaryMessage,
			expectedEncoding: binding.EncodingBinary,
			expectedKey:      true,
		},
		{
			name:             "Event to Structured - with key",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			addPartitionKey:  true,
			messageFactory:   func(e event.Event) binding.Message { return (*binding.EventMessage)(&e) },
			expectedEncoding: binding.EncodingStructured,
			expectedKey:      true,
		},
		{
			name:             "Event to Binary - with key",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary),
			addPartitionKey:  true,
			messageFactory:   func(e event.Event) binding.Message { return (*binding.EventMessage)(&e) },
			expectedEncoding: binding.EncodingBinary,
			expectedKey:      true,
		},
	}
	EachEvent(t, Events(), func(t *testing.T, e event.Event) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx := tt.context

				kafkaMessage := &sarama.ProducerMessage{
					Topic: "aaa",
				}

				eventIn := ConvertEventExtensionsToString(t, e.Clone())
				if tt.addPartitionKey {
					eventIn.SetExtension(partitionKey, testKey)
				}
				messageIn := tt.messageFactory(eventIn)

				err := WriteProducerMessage(ctx, messageIn, kafkaMessage)
				require.NoError(t, err)

				//Little hack to go back to Message
				headers := make(map[string][]byte)
				for _, h := range kafkaMessage.Headers {
					headers[strings.ToLower(string(h.Key))] = h.Value
				}

				var value []byte
				if kafkaMessage.Value != nil {
					value, err = kafkaMessage.Value.Encode()
					require.NoError(t, err)
				}

				messageOut := NewMessage(value, string(headers[contentTypeHeader]), headers)
				require.Equal(t, tt.expectedEncoding, messageOut.ReadEncoding())

				eventOut, err := binding.ToEvent(context.TODO(), messageOut)
				require.NoError(t, err)
				AssertEventEquals(t, eventIn, *eventOut)

				if !tt.expectedKey {
					require.Nil(t, kafkaMessage.Key)
				} else {
					require.NotNil(t, kafkaMessage.Key)
					val, err := kafkaMessage.Key.Encode()
					require.NoError(t, err)
					require.Equal(t, testKey, string(val))
				}
			})
		}
	})

}
