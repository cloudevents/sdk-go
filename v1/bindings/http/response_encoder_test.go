package http

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v1"
	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/test"
	ce "github.com/cloudevents/sdk-go/v1/cloudevents"
)

func TestEncodeHttpResponse(t *testing.T) {
	tests := []struct {
		name             string
		context          context.Context
		messageFactory   func(e cloudevents.Event) binding.Message
		expectedEncoding binding.Encoding
	}{
		{
			name:             "Structured to Structured",
			context:          context.TODO(),
			messageFactory:   test.NewMockStructuredMessage,
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Binary to Binary",
			context:          context.TODO(),
			messageFactory:   test.NewMockBinaryMessage,
			expectedEncoding: binding.EncodingBinary,
		},
		{
			name:             "Event to Structured",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			messageFactory:   func(e cloudevents.Event) binding.Message { return binding.EventMessage(e) },
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Event to Binary",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary),
			messageFactory:   func(e cloudevents.Event) binding.Message { return binding.EventMessage(e) },
			expectedEncoding: binding.EncodingBinary,
		},
	}
	for _, tt := range tests {
		test.EachEvent(t, test.Events(), func(t *testing.T, eventIn ce.Event) {
			t.Run(tt.name, func(t *testing.T) {
				res := &http.Response{
					Header: make(http.Header),
				}

				eventIn = test.ExToStr(t, eventIn)
				messageIn := tt.messageFactory(eventIn)

				err := EncodeHttpResponse(tt.context, messageIn, res, binding.TransformerFactories{})
				require.NoError(t, err)

				//Little hack to go back to Message
				messageOut, err := NewMessage(res.Header, res.Body)
				require.NoError(t, err)

				eventOut, encoding, err := binding.ToEvent(context.TODO(), messageOut)
				require.Equal(t, encoding, tt.expectedEncoding)
				test.AssertEventEquals(t, eventIn, eventOut)
			})
		})
	}
}
