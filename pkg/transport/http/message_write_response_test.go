package http

import (
	"context"
	"net/http"
	"testing"

	test2 "github.com/cloudevents/sdk-go/pkg/binding/test"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/event"
)

func TestEncodeHttpResponse(t *testing.T) {
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
			messageFactory:   func(e event.Event) binding.Message { return binding.NewEventMessage(&e) },
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Event to Binary",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary),
			messageFactory:   func(e event.Event) binding.Message { return binding.NewEventMessage(&e) },
			expectedEncoding: binding.EncodingBinary,
		},
	}
	for _, tt := range tests {
		test2.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
			t.Run(tt.name, func(t *testing.T) {
				res := &http.Response{
					Header: make(http.Header),
				}

				eventIn = test.CopyEventContext(test2.ExToStr(t, eventIn))
				messageIn := tt.messageFactory(eventIn)

				err := WriteHttpResponse(tt.context, messageIn, res, nil)
				require.NoError(t, err)

				//Little hack to go back to Message
				messageOut := NewMessageFromHttpResponse(res)
				require.Equal(t, tt.expectedEncoding, messageOut.ReadEncoding())

				eventOut, err := binding.ToEvent(context.TODO(), messageOut, nil)
				test2.AssertEventEquals(t, eventIn, *eventOut)
			})
		})
	}
}
