package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/event"
	nethttp "net/http"
	"net/url"
	"testing"

	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/types"
	"github.com/google/go-cmp/cmp"
)

func strptr(s string) *string {
	return &s
}

func TestDefaultBinaryEncodingSelectionStrategy(t *testing.T) {
	testCases := map[string]struct {
		event event.Event
		want  http.Encoding
	}{
		"default, unknown version": {
			event: event.Event{
				Context: &event.EventContextV1{
					SpecVersion: "unknown",
				},
			},
			want: http.Default,
		},
		"v0.3": {
			event: event.Event{
				Context: event.EventContextV03{}.AsV03(),
			},
			want: http.BinaryV03,
		},
		"v1.0": {
			event: event.Event{
				Context: event.EventContextV1{}.AsV1(),
			},
			want: http.BinaryV1,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := http.DefaultBinaryEncodingSelectionStrategy(context.TODO(), tc.event)

			if got != tc.want {
				t.Errorf("unexpected selection want: %s, got: %s", tc.want, got)
			}
		})
	}
}

func TestDefaultStructuredEncodingSelectionStrategy(t *testing.T) {
	testCases := map[string]struct {
		event event.Event
		want  http.Encoding
	}{
		"default, unknown version": {
			event: event.Event{
				Context: &event.EventContextV1{
					SpecVersion: "unknown",
				},
			},
			want: http.Default,
		},
		"v0.3": {
			event: event.Event{
				Context: event.EventContextV03{}.AsV03(),
			},
			want: http.StructuredV03,
		},
		"v1.0": {
			event: event.Event{
				Context: event.EventContextV1{}.AsV1(),
			},
			want: http.StructuredV1,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := http.DefaultStructuredEncodingSelectionStrategy(context.TODO(), tc.event)

			if got != tc.want {
				t.Errorf("unexpected selection want: %s, got: %s", tc.want, got)
			}
		})
	}
}

func TestCodecEncode(t *testing.T) {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}
	sourceUri := &types.URIRef{URL: *sourceUrl}

	testCases := map[string]struct {
		codec   *http.Codec
		event   event.Event
		want    *http.Message
		wantErr error
	}{
		"default v0.3 binary": {
			codec: &http.Codec{
				DefaultEncodingSelectionFn: http.DefaultBinaryEncodingSelectionStrategy,
			},
			event: event.Event{
				Context: event.EventContextV03{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				}.AsV03(),
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.3"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
				},
			},
		},
		"default v0.3 structured": {
			codec: &http.Codec{
				DefaultEncodingSelectionFn: http.DefaultStructuredEncodingSelectionStrategy,
			},
			event: event.Event{
				Context: event.EventContextV03{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				}.AsV03(),
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "0.3",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"simple v0.3 binary": {
			codec: &http.Codec{Encoding: http.BinaryV03},
			event: event.Event{
				Context: event.EventContextV03{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				}.AsV03(),
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.3"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
				},
			},
		},
		"simple v0.3 structured": {
			codec: &http.Codec{Encoding: http.StructuredV03},
			event: event.Event{
				Context: event.EventContextV03{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				}.AsV03(),
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "0.3",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"default v1.0 binary": {
			codec: &http.Codec{
				DefaultEncodingSelectionFn: http.DefaultBinaryEncodingSelectionStrategy,
			},
			event: event.Event{
				Context: event.EventContextV1{
					Type:   "com.example.test",
					Source: *sourceUri,
					ID:     "ABC-123",
				}.AsV1(),
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"1.0"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
				},
			},
		},
		"default v1.0 structured": {
			codec: &http.Codec{
				DefaultEncodingSelectionFn: http.DefaultStructuredEncodingSelectionStrategy,
			},
			event: event.Event{
				Context: event.EventContextV1{
					Type:   "com.example.test",
					Source: *sourceUri,
					ID:     "ABC-123",
				}.AsV1(),
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "1.0",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"simple v1.0 binary": {
			codec: &http.Codec{Encoding: http.BinaryV1},
			event: event.Event{
				Context: event.EventContextV1{
					Type:   "com.example.test",
					Source: *sourceUri,
					ID:     "ABC-123",
				}.AsV1(),
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"1.0"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
				},
			},
		},
		"simple v1.0 structured": {
			codec: &http.Codec{Encoding: http.StructuredV1},
			event: event.Event{
				Context: event.EventContextV1{
					Type:   "com.example.test",
					Source: *sourceUri,
					ID:     "ABC-123",
				}.AsV1(),
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "1.0",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, err := tc.codec.Encode(context.TODO(), tc.event)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {

				if msg, ok := got.(*http.Message); ok {
					// It is hard to read the byte dump
					want := string(tc.want.Body)
					got := string(msg.Body)
					if diff := cmp.Diff(want, got); diff != "" {
						t.Errorf("unexpected (-want, +got) = %v", diff)
						return
					}
				}
				t.Errorf("unexpected message (-want, +got) = %v", diff)
			}
		})
	}
}

// A cmp.Transformer to normalize case of http.Header map keys.
var normalizeHeaders = cmp.Transformer("NormalizeHeaders",
	func(in nethttp.Header) nethttp.Header {
		out := nethttp.Header{}
		for k, v := range in {
			out[nethttp.CanonicalHeaderKey(k)] = v
		}
		return out
	})

func TestCodecDecode(t *testing.T) {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}
	sourceUri := &types.URIRef{URL: *sourceUrl}

	testCases := map[string]struct {
		codec   *http.Codec
		msg     *http.Message
		want    *event.Event
		wantErr error
	}{
		"simple v0.3 binary": {
			codec: &http.Codec{Encoding: http.BinaryV03},
			msg: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.3"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
			want: &event.Event{
				Context: &event.EventContextV03{
					SpecVersion:     event.CloudEventsVersionV03,
					Type:            "com.example.test",
					Source:          *source,
					ID:              "ABC-123",
					DataContentType: event.StringOfApplicationJSON(),
				},
			},
		},
		"simple v0.3 structured": {
			codec: &http.Codec{Encoding: http.StructuredV03},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "0.3",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
			want: &event.Event{
				Context: &event.EventContextV03{
					SpecVersion: event.CloudEventsVersionV03,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
				},
			},
		},
		"simple v1.0 binary": {
			codec: &http.Codec{Encoding: http.BinaryV1},
			msg: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"1.0"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
			want: &event.Event{
				Context: &event.EventContextV1{
					SpecVersion:     event.CloudEventsVersionV1,
					Type:            "com.example.test",
					Source:          *sourceUri,
					ID:              "ABC-123",
					DataContentType: event.StringOfApplicationJSON(),
				},
			},
		},
		"simple v1.0 structured": {
			codec: &http.Codec{Encoding: http.StructuredV1},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "1.0",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
			want: &event.Event{
				Context: &event.EventContextV1{
					SpecVersion: event.CloudEventsVersionV1,
					Type:        "com.example.test",
					Source:      *sourceUri,
					ID:          "ABC-123",
				},
			},
		},

		// Conversion tests.

		"simple v1.0 binary -> v0.3 binary": {
			codec: &http.Codec{Encoding: http.BinaryV03},
			msg: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"1.0"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
			want: &event.Event{
				Context: &event.EventContextV03{
					SpecVersion:     event.CloudEventsVersionV03,
					Type:            "com.example.test",
					Source:          *source,
					ID:              "ABC-123",
					DataContentType: event.StringOfApplicationJSON(),
				},
			},
		},
		"simple v1.0 structured -> v0.3 structured": {
			codec: &http.Codec{Encoding: http.StructuredV03},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "1.0",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
			want: &event.Event{
				Context: &event.EventContextV03{
					SpecVersion: event.CloudEventsVersionV03,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
				},
			},
		},
		"simple v0.3 binary -> v1.0 binary": {
			codec: &http.Codec{Encoding: http.BinaryV1},
			msg: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.3"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
			want: &event.Event{
				Context: &event.EventContextV1{
					SpecVersion:     event.CloudEventsVersionV1,
					Type:            "com.example.test",
					Source:          *sourceUri,
					ID:              "ABC-123",
					DataContentType: event.StringOfApplicationJSON(),
				},
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, err := tc.codec.Decode(context.TODO(), tc.msg)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected event (-want, +got) = %v", diff)
			}
			// Round trip thru a http.Request
			var req nethttp.Request
			tc.msg.ToRequest(&req)
			gotm, err := http.NewMessage(req.Header, req.Body)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(tc.msg, gotm, normalizeHeaders); diff != "" {
				t.Errorf("unexpected message (-want, +got) = %v", diff)
			}
		})
	}
}

func toBytes(body map[string]interface{}) []byte {
	b, err := json.Marshal(body)
	if err != nil {
		return []byte(fmt.Sprintf(`{"error":%q}`, err.Error()))
	}
	return b
}
