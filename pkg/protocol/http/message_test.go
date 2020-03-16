package http

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/event"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name     string
		encoding binding.Encoding
	}{
		{
			name:     "Structured encoding",
			encoding: binding.EncodingStructured,
		},
		{
			name:     "Binary encoding",
			encoding: binding.EncodingBinary,
		},
	}
	for _, tt := range tests {
		test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
			t.Run(tt.name, func(t *testing.T) {

				ctx := context.TODO()
				if tt.encoding == binding.EncodingStructured {
					ctx = binding.WithForceStructured(ctx)
				} else if tt.encoding == binding.EncodingBinary {
					ctx = binding.WithForceBinary(ctx)
				}

				req := httptest.NewRequest("POST", "http://localhost", nil)
				require.NoError(t, WriteRequest(ctx, (*binding.EventMessage)(&eventIn), req, binding.TransformerFactories{}))

				got := NewMessageFromHttpRequest(req)
				require.Equal(t, tt.encoding, got.ReadEncoding())
			})
		})
	}
}

func TestNewMessageUnknown(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		req := httptest.NewRequest("POST", "http://localhost", bytes.NewReader([]byte("{}")))
		req.Header.Add("content-type", "application/json")

		got := NewMessageFromHttpRequest(req)
		require.Equal(t, binding.EncodingUnknown, got.ReadEncoding())
	})
}
