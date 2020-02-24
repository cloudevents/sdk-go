// +build kafka

package kafka_sarama

import (
	"context"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

func TestEncodeKafkaProducerMessage(t *testing.T) {
	tests := []struct {
		name             string
		context          context.Context
		messageFactory   func(e cloudevents.Event) binding.Message
		expectedEncoding binding.Encoding
		skipKey          bool
	}{
		{
			name:             "Structured to Structured with Skip key",
			context:          context.TODO(),
			messageFactory:   func(e cloudevents.Event) binding.Message { return test.NewMockStructuredMessage(e) },
			expectedEncoding: binding.EncodingStructured,
			skipKey:          true,
		},
		{
			name:             "Binary to Binary with Skip key",
			context:          context.TODO(),
			messageFactory:   func(e cloudevents.Event) binding.Message { return test.NewMockBinaryMessage(e) },
			expectedEncoding: binding.EncodingBinary,
			skipKey:          true,
		},
		{
			name:             "Event to Structured with Skip key",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			messageFactory:   func(e cloudevents.Event) binding.Message { return binding.EventMessage(e) },
			expectedEncoding: binding.EncodingStructured,
			skipKey:          true,
		},
		{
			name:             "Event to Binary with Skip key",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary),
			messageFactory:   func(e cloudevents.Event) binding.Message { return binding.EventMessage(e) },
			expectedEncoding: binding.EncodingBinary,
			skipKey:          true,
		},
		{
			name:             "Structured to Structured",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			messageFactory:   func(e cloudevents.Event) binding.Message { return test.NewMockStructuredMessage(e) },
			expectedEncoding: binding.EncodingStructured,
			skipKey:          false,
		},
		{
			name:             "Binary to Binary",
			context:          context.TODO(),
			messageFactory:   func(e cloudevents.Event) binding.Message { return test.NewMockBinaryMessage(e) },
			expectedEncoding: binding.EncodingBinary,
			skipKey:          false,
		},
		{
			name:             "Event to Structured",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			messageFactory:   func(e cloudevents.Event) binding.Message { return binding.EventMessage(e) },
			expectedEncoding: binding.EncodingStructured,
			skipKey:          false,
		},
		{
			name:             "Event to Binary",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary),
			messageFactory:   func(e cloudevents.Event) binding.Message { return binding.EventMessage(e) },
			expectedEncoding: binding.EncodingBinary,
			skipKey:          false,
		},
	}
	for _, tt := range tests {
		test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
			t.Run(tt.name, func(t *testing.T) {
				ctx := tt.context

				if tt.skipKey {
					ctx = WithSkipKeyExtension(ctx)
				}

				message := &sarama.ProducerMessage{
					Topic: "aaa",
				}

				eventIn = test.ExToStr(t, eventIn)
				messageIn := tt.messageFactory(eventIn)

				err := EncodeKafkaProducerMessage(ctx, messageIn, message, binding.TransformerFactories{})
				require.NoError(t, err)

				//Little hack to go back to Message
				headers := make(map[string][]byte)
				for _, h := range message.Headers {
					headers[string(h.Key)] = h.Value
				}

				var key []byte
				if message.Key != nil {
					key, err = message.Key.Encode()
					require.NoError(t, err)
				}

				var value []byte
				if message.Value != nil {
					value, err = message.Value.Encode()
					require.NoError(t, err)
				}

				messageOut, err := NewMessageFromRaw(key, value, string(headers[ContentType]), headers)
				require.NoError(t, err)

				eventOut, encoding, err := binding.ToEvent(context.TODO(), messageOut)
				require.NoError(t, err)
				require.Equal(t, tt.expectedEncoding, encoding)
				test.AssertEventEquals(t, eventIn, eventOut)
			})
		})
	}
}
