package http

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
)

func TestNewMessageFromHttpRequest(t *testing.T) {
	tests := []struct {
		name     string
		encoding binding.Encoding
	}{{
		name:     "Structured encoding",
		encoding: binding.EncodingStructured,
	}, {
		name:     "Binary encoding",
		encoding: binding.EncodingBinary,
	}}
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
				require.NoError(t, WriteRequest(ctx, (*binding.EventMessage)(&eventIn), req))

				got := NewMessageFromHttpRequest(req)
				require.Equal(t, tt.encoding, got.ReadEncoding())

				require.NoError(t, got.Finish(nil))
			})
		})
	}
}

func TestNewMessageFromHttpRequestUnknown(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		req := httptest.NewRequest("POST", "http://localhost", bytes.NewReader([]byte("{}")))
		req.Header.Add("content-type", "application/json")

		got := NewMessageFromHttpRequest(req)

		require.Equal(t, binding.EncodingUnknown, got.ReadEncoding())
		require.NoError(t, got.Finish(nil))
	})
}

func TestNewMessageFromHttpRequestWithOnFinish(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		req := httptest.NewRequest("POST", "http://localhost", bytes.NewReader([]byte("{}")))
		req.Header.Add("content-type", "application/json")

		got := NewMessageFromHttpRequest(req)
		require.Equal(t, binding.EncodingUnknown, got.ReadEncoding())

		// Just no-op the error.
		got.OnFinish = func(err error) error {
			return err
		}

		require.NoError(t, got.Finish(nil))
		require.Error(t, got.Finish(errors.New("unit test")))
	})
}

func TestNewMessageFromHttpRequestWithOnFinish_errors(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		req := httptest.NewRequest("POST", "http://localhost", bytes.NewReader([]byte("{}")))
		req.Header.Add("content-type", "application/json")

		got := NewMessageFromHttpRequest(req)
		require.Equal(t, binding.EncodingUnknown, got.ReadEncoding())

		// Just no-op the error.
		got.OnFinish = func(err error) error {
			return errors.New("unit test")
		}

		require.Error(t, got.Finish(nil))
		require.Error(t, got.Finish(errors.New("unit test")))
	})
}

func TestNewMessageFromHttpRequestNoBody(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		req := httptest.NewRequest("POST", "http://localhost", nil)
		req.Header.Add("content-type", "application/json")

		got := NewMessageFromHttpRequest(req)
		require.Equal(t, binding.EncodingUnknown, got.ReadEncoding())

		require.NoError(t, got.Finish(nil))
	})
}

func TestMessageMetadataReader(t *testing.T) {
	eventIn := test.FullEvent()
	req := httptest.NewRequest("POST", "http://localhost", nil)
	require.NoError(t, WriteRequest(binding.WithForceBinary(context.TODO()), (*binding.EventMessage)(&eventIn), req))

	got := binding.MessageMetadataReader(NewMessageFromHttpRequest(req))
	require.Equal(t, eventIn.Extensions()["exstring"], got.GetExtension("exstring"))
	_, id := got.GetAttribute(spec.ID)
	require.Equal(t, eventIn.ID(), id)
}

func TestMessageTransformDeleteExtension(t *testing.T) {
	eventIn := test.FullEvent()
	req := httptest.NewRequest("POST", "http://localhost", nil)
	msg := bindingtest.MustCreateMockBinaryMessage(eventIn)
	require.NoError(t, WriteRequest(binding.WithForceBinary(context.TODO()), msg, req, transformer.DeleteExtension("exstring")))

	got := binding.MessageMetadataReader(NewMessageFromHttpRequest(req))
	require.Equal(t, nil, got.GetExtension("exstring"))
	_, id := got.GetAttribute(spec.ID)
	require.Equal(t, eventIn.ID(), id)
}

func TestNewMessageFromHttpResponse(t *testing.T) {
	tests := []struct {
		name     string
		encoding binding.Encoding
		resp     *http.Response
	}{{
		name:     "Structured encoding",
		encoding: binding.EncodingStructured,
		resp: &http.Response{
			Header: http.Header{
				"Content-Type": {event.ApplicationCloudEventsJSON},
			},
			Body:          ioutil.NopCloser(bytes.NewReader([]byte(`{"data":"foo","datacontenttype":"application/json","id":"id","source":"source","specversion":"1.0","type":"type"}`))),
			ContentLength: 113,
		},
	}, {
		name:     "Binary encoding",
		encoding: binding.EncodingBinary,
		resp: &http.Response{
			Header: func() http.Header {
				h := http.Header{}
				h.Set("ce-specversion", "1.0")
				h.Set("ce-source", "unittest")
				h.Set("ce-type", "unittest")
				h.Set("ce-id", "unittest")
				h.Set("Content-Type", "application/json")
				return h
			}(),
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMessageFromHttpResponse(tt.resp)
			require.Equal(t, tt.encoding, got.ReadEncoding())

			require.NoError(t, got.Finish(nil))
		})
	}
}
