package http

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/buffering"
	"github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
)

func TestWriteHttpResponseWriter(t *testing.T) {
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
			name:    "Structured to buffered to Structured",
			context: context.TODO(),
			messageFactory: func(e event.Event) binding.Message {
				m := test.MustCreateMockStructuredMessage(e)

				buffered, err := buffering.BufferMessage(context.TODO(), m)
				require.NoError(t, err)

				return buffered
			},
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:    "Binary to buffered to Binary",
			context: context.TODO(),
			messageFactory: func(e event.Event) binding.Message {
				m := test.MustCreateMockBinaryMessage(e)

				buffered, err := buffering.BufferMessage(context.TODO(), m)
				require.NoError(t, err)

				return buffered
			},
			expectedEncoding: binding.EncodingBinary,
		},
		{
			name:    "Direct structured HttpRequest to Structured",
			context: context.TODO(),
			messageFactory: func(e event.Event) binding.Message {
				req := httptest.NewRequest("POST", "http://localhost", nil)
				require.NoError(t, WriteRequest(binding.WithForceStructured(context.TODO()), binding.ToMessage(&e), req))

				return NewMessageFromHttpRequest(req)
			},
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:    "Binary to buffered to Binary",
			context: context.TODO(),
			messageFactory: func(e event.Event) binding.Message {
				req := httptest.NewRequest("POST", "http://localhost", nil)
				require.NoError(t, WriteRequest(binding.WithForceBinary(context.TODO()), binding.ToMessage(&e), req))

				return NewMessageFromHttpRequest(req)
			},
			expectedEncoding: binding.EncodingBinary,
		},
		{
			name:             "Event to Structured",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			messageFactory:   func(e event.Event) binding.Message { return binding.ToMessage(&e) },
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:             "Event to Binary",
			context:          binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary),
			messageFactory:   func(e event.Event) binding.Message { return binding.ToMessage(&e) },
			expectedEncoding: binding.EncodingBinary,
		},
	}
	for _, tt := range tests {
		test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
			t.Run(tt.name, func(t *testing.T) {
				res := httptest.NewRecorder()

				eventIn = test.ExToStr(t, eventIn)
				messageIn := tt.messageFactory(eventIn)

				shouldHaveContentLength := eventIn.Data() != nil || messageIn.ReadEncoding() == binding.EncodingStructured

				err := WriteResponseWriter(tt.context, messageIn, 200, res)
				require.NoError(t, err)

				require.Equal(t, 200, res.Code)
				if shouldHaveContentLength {
					require.NotZero(t, res.Header().Get("content-length"))
				}

				//Little hack to go back to Message
				messageOut := NewMessage(res.Header(), ioutil.NopCloser(bytes.NewReader(res.Body.Bytes())))
				require.Equal(t, tt.expectedEncoding, messageOut.ReadEncoding())

				eventOut, err := binding.ToEvent(context.TODO(), messageOut)
				require.NoError(t, err)
				test.AssertEventEquals(t, eventIn, *eventOut)
			})
		})
	}
}
