package http

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/buffering"
	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
)

func TestWriteHttpResponseWriter(t *testing.T) {
	tests := []struct {
		name                string
		context             context.Context
		messageFactory      func(e event.Event) binding.Message
		expectedEncoding    binding.Encoding
		expectContentLength bool
	}{
		{
			name:    "Structured to Structured",
			context: context.TODO(),
			messageFactory: func(e event.Event) binding.Message {
				return bindingtest.MustCreateMockStructuredMessage(t, e)
			},
			expectedEncoding:    binding.EncodingStructured,
			expectContentLength: true,
		},
		{
			name:                "Binary to Binary",
			context:             context.TODO(),
			messageFactory:      bindingtest.MustCreateMockBinaryMessage,
			expectedEncoding:    binding.EncodingBinary,
			expectContentLength: true,
		},
		{
			name:    "Structured to buffered to Structured",
			context: context.TODO(),
			messageFactory: func(e event.Event) binding.Message {
				m := bindingtest.MustCreateMockStructuredMessage(t, e)

				buffered, err := buffering.BufferMessage(context.TODO(), m)
				require.NoError(t, err)

				return buffered
			},
			expectedEncoding:    binding.EncodingStructured,
			expectContentLength: true,
		},
		{
			name:    "Binary to buffered to Binary",
			context: context.TODO(),
			messageFactory: func(e event.Event) binding.Message {
				m := bindingtest.MustCreateMockBinaryMessage(e)

				buffered, err := buffering.BufferMessage(context.TODO(), m)
				require.NoError(t, err)

				return buffered
			},
			expectedEncoding:    binding.EncodingBinary,
			expectContentLength: true,
		},
		{
			name:    "Direct structured HttpRequest to Structured",
			context: context.TODO(),
			messageFactory: func(e event.Event) binding.Message {
				req := httptest.NewRequest("POST", "http://localhost", nil)
				require.NoError(t, WriteRequest(binding.WithForceStructured(context.TODO()), binding.ToMessage(&e), req))

				return NewMessageFromHttpRequest(req)
			},
			expectedEncoding:    binding.EncodingStructured,
			expectContentLength: false,
		},
		{
			name:    "Binary to binary HttpRequest to Binary",
			context: context.TODO(),
			messageFactory: func(e event.Event) binding.Message {
				req := httptest.NewRequest("POST", "http://localhost", nil)
				require.NoError(t, WriteRequest(binding.WithForceBinary(context.TODO()), binding.ToMessage(&e), req))

				return NewMessageFromHttpRequest(req)
			},
			expectedEncoding:    binding.EncodingBinary,
			expectContentLength: false,
		},
		{
			name:                "Event to Structured",
			context:             binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingStructured),
			messageFactory:      func(e event.Event) binding.Message { return binding.ToMessage(&e) },
			expectedEncoding:    binding.EncodingStructured,
			expectContentLength: true,
		},
		{
			name:                "Event to Binary",
			context:             binding.WithPreferredEventEncoding(context.TODO(), binding.EncodingBinary),
			messageFactory:      func(e event.Event) binding.Message { return binding.ToMessage(&e) },
			expectedEncoding:    binding.EncodingBinary,
			expectContentLength: true,
		},
	}
	for _, tt := range tests {
		test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
			t.Run(tt.name, func(t *testing.T) {
				res := httptest.NewRecorder()

				eventIn = test.ConvertEventExtensionsToString(t, eventIn)
				messageIn := tt.messageFactory(eventIn)

				shouldHaveContentLength := tt.expectContentLength && (eventIn.Data() != nil || messageIn.ReadEncoding() == binding.EncodingStructured)

				err := WriteResponseWriter(tt.context, messageIn, 202, res)
				require.NoError(t, err)

				response := res.Result()

				require.Equal(t, 202, response.StatusCode)
				if shouldHaveContentLength {
					require.NotZero(t, response.Header.Get("content-length"))
				}

				//Little hack to go back to Message
				messageOut := NewMessageFromHttpResponse(response)
				require.Equal(t, tt.expectedEncoding, messageOut.ReadEncoding())

				eventOut, err := binding.ToEvent(context.TODO(), messageOut)
				require.NoError(t, err)
				test.AssertEventEquals(t, eventIn, *eventOut)
			})
		})
	}
}

func TestWriteHttpResponseWriter_using_transformers_with_end(t *testing.T) {
	eventIn := test.ConvertEventExtensionsToString(t, test.FullEvent())
	initialReq := httptest.NewRequest("POST", "http://localhost", nil)
	require.NoError(t, WriteRequest(binding.WithForceBinary(context.TODO()), binding.ToMessage(&eventIn), initialReq))

	messageIn := NewMessageFromHttpRequest(initialReq)

	res := httptest.NewRecorder()

	err := WriteResponseWriter(context.TODO(), messageIn, 202, res, transformer.AddExtension("blablabla", "blablabla"))
	require.NoError(t, err)

	response := res.Result()

	require.Equal(t, 202, response.StatusCode)

	//Little hack to go back to Message
	messageOut := NewMessageFromHttpResponse(response)
	require.Equal(t, binding.EncodingBinary, messageOut.ReadEncoding())

	eventIn.SetExtension("blablabla", "blablabla")
	eventOut, err := binding.ToEvent(context.TODO(), messageOut)
	require.NoError(t, err)
	test.AssertEventEquals(t, eventIn, *eventOut)
}

func TestWriteHttpResponseWriter_using_transformers_fails(t *testing.T) {
	eventIn := test.ConvertEventExtensionsToString(t, test.FullEvent())
	messageIn := bindingtest.MustCreateMockBinaryMessage(eventIn)
	messageIn.(*bindingtest.MockBinaryMessage).Extensions["badext"] = struct {
		val string
	}{
		val: "aaa",
	}

	res := httptest.NewRecorder()

	err := WriteResponseWriter(context.TODO(), messageIn, 202, res)
	require.Error(t, err)
	res.WriteHeader(500)

	response := res.Result()

	require.Equal(t, 500, response.StatusCode)
}
