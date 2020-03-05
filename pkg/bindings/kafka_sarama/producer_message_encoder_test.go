package kafka_sarama

import (
	"context"
	"strings"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/event"
)

func TestEncodeKafkaProducerMessage(t *testing.T) {
	tests := []struct {
		name             string
		context          context.Context
		messageFactory   func(e event.Event) binding.Message
		expectedEncoding binding.Encoding
		skipKey          bool
	}{
		{
			name:             "Structured to Structured with Skip key",
			context:          context.TODO(),
			messageFactory:   func(e event.Event) binding.Message { return test.MustCreateMockStructuredMessage(e) },
			expectedEncoding: binding.EncodingStructured,
			skipKey:          true,
		},
		{
			name:             "Binary to Binary with Skip key",
			context:          context.TODO(),
			messageFactory:   func(e event.Event) binding.Message { return test.MustCreateMockBinaryMessage(e) },
			expectedEncoding: binding.EncodingBinary,
			skipKey:          true,
		},
		{
			name:             "Event to Structured with Skip key",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			messageFactory:   func(e event.Event) binding.Message { return binding.EventMessage(e) },
			expectedEncoding: binding.EncodingStructured,
			skipKey:          true,
		},
		{
			name:             "Event to Binary with Skip key",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary),
			messageFactory:   func(e event.Event) binding.Message { return binding.EventMessage(e) },
			expectedEncoding: binding.EncodingBinary,
			skipKey:          true,
		},
		{
			name:             "Structured to Structured",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			messageFactory:   func(e event.Event) binding.Message { return test.MustCreateMockStructuredMessage(e) },
			expectedEncoding: binding.EncodingEvent,
			skipKey:          false,
		},
		{
			name:             "Binary to Binary",
			context:          context.TODO(),
			messageFactory:   func(e event.Event) binding.Message { return test.MustCreateMockBinaryMessage(e) },
			expectedEncoding: binding.EncodingBinary,
			skipKey:          false,
		},
		{
			name:             "Event to Structured",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			messageFactory:   func(e event.Event) binding.Message { return binding.EventMessage(e) },
			expectedEncoding: binding.EncodingEvent,
			skipKey:          false,
		},
		{
			name:             "Event to Binary",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary),
			messageFactory:   func(e event.Event) binding.Message { return binding.EventMessage(e) },
			expectedEncoding: binding.EncodingBinary,
			skipKey:          false,
		},
	}
	for _, tt := range tests {
		test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
			t.Run(tt.name, func(t *testing.T) {
				ctx := tt.context

				if tt.skipKey {
					ctx = WithSkipKeyExtension(ctx)
				} else {
					eventIn.SetExtension("key", "bla")
				}

				kafkaMessage := &sarama.ProducerMessage{
					Topic: "aaa",
				}

				eventIn = test.ExToStr(t, eventIn)
				messageIn := tt.messageFactory(eventIn)

				err := EncodeKafkaProducerMessage(ctx, messageIn, kafkaMessage, nil)
				require.NoError(t, err)

				//Little hack to go back to Message
				headers := make(map[string][]byte)
				for _, h := range kafkaMessage.Headers {
					headers[strings.ToLower(string(h.Key))] = h.Value
				}

				var key []byte
				if kafkaMessage.Key != nil {
					key, err = kafkaMessage.Key.Encode()
					require.NoError(t, err)
				}

				var value []byte
				if kafkaMessage.Value != nil {
					value, err = kafkaMessage.Value.Encode()
					require.NoError(t, err)
				}

				if !tt.skipKey {
					require.Equal(t, []byte("bla"), key)
				}

				messageOut, err := NewMessage(key, value, string(headers[contentTypeHeader]), headers)
				require.NoError(t, err)

				eventOut, encoding, err := binding.ToEvent(context.TODO(), messageOut, nil)
				require.NoError(t, err)
				require.Equal(t, tt.expectedEncoding, encoding)
				test.AssertEventEquals(t, eventIn, eventOut)
			})
		})
	}
}
