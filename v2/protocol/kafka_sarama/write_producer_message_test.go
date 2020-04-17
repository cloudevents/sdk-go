package kafka_sarama

import (
	"context"
	"strings"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
)

func TestEncodeKafkaProducerMessage(t *testing.T) {
	tests := []struct {
		name             string
		context          context.Context
		messageFactory   func(e event.Event) binding.Message
		expectedEncoding binding.Encoding
	}{
		{
			name:             "Structured to Structured",
			context:          context.TODO(),
			messageFactory:   test.MustCreateMockStructuredMessage,
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Binary to Binary",
			context:          context.TODO(),
			messageFactory:   test.MustCreateMockBinaryMessage,
			expectedEncoding: binding.EncodingBinary,
		},
		{
			name:             "Event to Structured",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			messageFactory:   func(e event.Event) binding.Message { return (*binding.EventMessage)(&e) },
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Event to Binary",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary),
			messageFactory:   func(e event.Event) binding.Message { return (*binding.EventMessage)(&e) },
			expectedEncoding: binding.EncodingBinary,
		},
	}
	for _, tt := range tests {
		test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
			t.Run(tt.name, func(t *testing.T) {
				ctx := tt.context

				kafkaMessage := &sarama.ProducerMessage{
					Topic: "aaa",
				}

				eventIn = test.ExToStr(t, eventIn)
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
				test.AssertEventEquals(t, eventIn, *eventOut)
			})
		})
	}
}
