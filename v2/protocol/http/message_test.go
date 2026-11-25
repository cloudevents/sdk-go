/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"bytes"
	"context"
	"errors"
	"io"
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
				switch tt.encoding {
				case binding.EncodingStructured:
					ctx = binding.WithForceStructured(ctx)
				case binding.EncodingBinary:
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

type testContextKey struct{}

func TestNewMessageFromHttpRequestContainsContext(t *testing.T) {
	eventIn := test.FullEvent()
	ctx := context.WithValue(context.Background(), testContextKey{}, "world")

	req := httptest.NewRequest("POST", "http://localhost", nil).WithContext(ctx)
	require.NoError(t, WriteRequest(ctx, (*binding.EventMessage)(&eventIn), req))

	got := NewMessageFromHttpRequest(req)

	require.Equal(t, ctx, got.Context())
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

func TestMessageTransformSetAttribute(t *testing.T) {
	eventIn := test.ConvertEventExtensionsToString(t, test.FullEvent())
	req := httptest.NewRequest("POST", "http://localhost", nil)
	msg := bindingtest.MustCreateMockBinaryMessage(eventIn)

	err := WriteRequest(
		binding.WithForceBinary(context.TODO()),
		msg,
		req,
		transformer.SetAttribute(spec.Type, func(interface{}) (interface{}, error) {
			return "updated.type", nil
		}),
		transformer.SetAttribute(spec.Subject, func(interface{}) (interface{}, error) {
			return "updated-subject", nil
		}),
		transformer.SetAttribute(spec.DataContentType, func(interface{}) (interface{}, error) {
			return "text/plain", nil
		}),
	)
	require.NoError(t, err)

	require.Equal(t, []string{"updated.type"}, req.Header.Values("Ce-Type"))
	require.Equal(t, []string{"updated-subject"}, req.Header.Values("Ce-Subject"))
	require.Equal(t, []string{"text/plain"}, req.Header.Values("Content-Type"))

	messageOut := NewMessageFromHttpRequest(req)
	_, gotType := messageOut.GetAttribute(spec.Type)
	require.Equal(t, "updated.type", gotType)
	_, gotSubject := messageOut.GetAttribute(spec.Subject)
	require.Equal(t, "updated-subject", gotSubject)
	_, gotDCT := messageOut.GetAttribute(spec.DataContentType)
	require.Equal(t, "text/plain", gotDCT)

	eventOut, err := binding.ToEvent(context.Background(), messageOut)
	require.NoError(t, err)
	require.Equal(t, "updated.type", eventOut.Type())
	require.Equal(t, "updated-subject", eventOut.Subject())
	require.Equal(t, "text/plain", eventOut.DataContentType())
}

// TestReadBinaryEmptyCeHeaderNoPanic verifies that a header whose name is
// exactly the "Ce-" prefix (i.e. no attribute/extension name follows it)
// does not cause ReadBinary to panic with an index-out-of-range error.
func TestReadBinaryEmptyCeHeaderNoPanic(t *testing.T) {
	h := http.Header{}
	h.Set("Ce-Specversion", "1.0")
	h.Set("Ce-Id", "1")
	h.Set("Ce-Source", "/")
	h.Set("Ce-Type", "t")
	h["Ce-"] = []string{"x"} // header name == prefix, no attribute/extension name

	m := NewMessage(h, nil)

	require.NotPanics(t, func() {
		_, err := binding.ToEvent(context.Background(), m)
		require.NoError(t, err)
	})
}

// TestReadBinaryCeDashDashHeaderNoPanic verifies that a header named "Ce--"
// (prefix "Ce-" followed by a single "-", yielding an invalid empty/dash
// extension name) does not cause ReadBinary to panic, and instead results
// in a validation error from the invalid extension name.
func TestReadBinaryCeDashDashHeaderNoPanic(t *testing.T) {
	h := http.Header{}
	h.Set("Ce-Specversion", "1.0")
	h.Set("Ce-Id", "1")
	h.Set("Ce-Source", "/")
	h.Set("Ce-Type", "t")
	h["Ce--"] = []string{"x"} // header name == prefix + "-", invalid extension name

	m := NewMessage(h, nil)

	require.NotPanics(t, func() {
		_, err := binding.ToEvent(context.Background(), m)
		require.Error(t, err)
		require.Contains(t, err.Error(), "bad key")
	})
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
			Body:          io.NopCloser(bytes.NewReader([]byte(`{"data":"foo","datacontenttype":"application/json","id":"id","source":"source","specversion":"1.0","type":"type"}`))),
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

func TestReadBinaryEncodedHeaders(t *testing.T) {
	for _, sv := range []string{"%31.0", `"1.0"`} {
		h := http.Header{}
		h.Set("Ce-Specversion", sv)
		h.Set("Ce-Id", "%41%20b")
		h.Set("Ce-Source", "/")
		h.Set("Ce-Type", "t")
		h.Set("Ce-Subject", `"x, y"`)
		h.Set("Ce-Myext", "%E2%82%AC")

		m := NewMessage(h, nil)
		require.Equal(t, binding.EncodingBinary, m.ReadEncoding(), sv)
		e, err := binding.ToEvent(context.Background(), m)
		require.NoError(t, err, sv)
		require.Equal(t, "1.0", e.SpecVersion())
		require.Equal(t, "A b", e.ID())
		require.Equal(t, "x, y", e.Subject())
		require.Equal(t, "€", e.Extensions()["myext"])
	}
}

func TestReadBinaryInvalidUTF8Header(t *testing.T) {
	for _, v := range []string{"%FF", "a\xffb"} {
		h := http.Header{}
		h.Set("Ce-Specversion", "1.0")
		h.Set("Ce-Id", "1")
		h.Set("Ce-Source", "/")
		h.Set("Ce-Type", "t")
		h.Set("Ce-Subject", v)

		_, err := binding.ToEvent(context.Background(), NewMessage(h, nil))
		require.Error(t, err, v)
		require.Contains(t, err.Error(), "invalid UTF-8")
	}
}

func TestWriteReadEncodedRoundTrip(t *testing.T) {
	in := event.New()
	in.SetID("100% \"q\" €")
	in.SetSource("/s")
	in.SetType("t")
	in.SetSubject("a,b  c")
	in.SetExtension("myext", "ü %41")

	req, err := http.NewRequest("POST", "http://localhost", nil)
	require.NoError(t, err)
	require.NoError(t, WriteRequest(context.Background(), binding.ToMessage(&in), req))
	require.Equal(t, "100%25 %22q%22 %E2%82%AC", req.Header.Get("Ce-Id"))

	out, err := binding.ToEvent(context.Background(), NewMessageFromHttpRequest(req))
	require.NoError(t, err)
	require.Equal(t, in.ID(), out.ID())
	require.Equal(t, in.Subject(), out.Subject())
	require.Equal(t, "ü %41", out.Extensions()["myext"])

	bad := event.New()
	bad.SetID("1")
	bad.SetSource("/s")
	bad.SetType("t")
	bad.SetSubject("a\xffb")
	req2, _ := http.NewRequest("POST", "http://localhost", nil)
	require.Error(t, WriteRequest(context.Background(), binding.ToMessage(&bad), req2))
}
