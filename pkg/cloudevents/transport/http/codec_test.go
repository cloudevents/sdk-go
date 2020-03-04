package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"net/url"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/pkg/event"

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
		"v0.2": {
			event: event.Event{
				Context: event.EventContextV02{}.AsV02(),
			},
			want: http.BinaryV02,
		},
		"v0.3": {
			event: event.Event{
				Context: event.EventContextV03{}.AsV03(),
			},
			want: http.BinaryV03,
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
		"v0.2": {
			event: event.Event{
				Context: event.EventContextV02{}.AsV02(),
			},
			want: http.StructuredV02,
		},
		"v0.3": {
			event: event.Event{
				Context: event.EventContextV03{}.AsV03(),
			},
			want: http.StructuredV03,
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

	testCases := map[string]struct {
		codec   *http.Codec
		event   event.Event
		want    *http.Message
		wantErr error
	}{
		"default v0.2 binary": {
			codec: &http.Codec{
				DefaultEncodingSelectionFn: http.DefaultBinaryEncodingSelectionStrategy,
			},
			event: event.Event{
				Context: event.EventContextV02{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				}.AsV02(),
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
				},
			},
		},
		"default v0.2 structured": {
			codec: &http.Codec{
				DefaultEncodingSelectionFn: http.DefaultStructuredEncodingSelectionStrategy,
			},
			event: event.Event{
				Context: event.EventContextV02{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				}.AsV02(),
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "0.2",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
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
		"simple v0.2 binary": {
			codec: &http.Codec{Encoding: http.BinaryV02},
			event: event.Event{
				Context: event.EventContextV02{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				}.AsV02(),
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
				},
			},
		},
		"simple v0.2 structured": {
			codec: &http.Codec{Encoding: http.StructuredV02},
			event: event.Event{
				Context: event.EventContextV02{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				}.AsV02(),
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "0.2",
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

	testCases := map[string]struct {
		codec   *http.Codec
		msg     *http.Message
		want    *event.Event
		wantErr error
	}{
		"simple v0.2 binary": {
			codec: &http.Codec{Encoding: http.BinaryV02},
			msg: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
			want: &event.Event{
				Context: &event.EventContextV02{
					SpecVersion: event.CloudEventsVersionV02,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
					ContentType: event.StringOfApplicationJSON(),
				},
			},
		},
		"simple v0.2 structured": {
			codec: &http.Codec{Encoding: http.StructuredV02},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "0.2",
						"id":          "ABC-123",
						"type":        "com.example.test",
						"source":      "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
			want: &event.Event{
				Context: &event.EventContextV02{
					SpecVersion: event.CloudEventsVersionV02,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
				},
			},
		},

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

		// Conversion tests.

		"simple v0.2 binary -> v0.3 binary": {
			codec: &http.Codec{Encoding: http.BinaryV03},
			msg: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
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
		"simple v0.2 structured -> v0.3 structured": {
			codec: &http.Codec{Encoding: http.StructuredV03},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion": "0.2",
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
		"simple v0.3 binary -> v0.2 binary": {
			codec: &http.Codec{Encoding: http.BinaryV02},
			msg: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
			want: &event.Event{
				Context: &event.EventContextV02{
					SpecVersion: event.CloudEventsVersionV02,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
					ContentType: event.StringOfApplicationJSON(),
				},
			},
		},
		// TODO:: add the v0.3 conversion tests. Might want to think of a new way to do this.
		// The tests are getting very long...
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

type DataExample struct {
	AnInt   int        `json:"a,omitempty"`
	AString string     `json:"b,omitempty"`
	AnArray []string   `json:"c,omitempty"`
	ATime   *time.Time `json:"e,omitempty"`
}

func TestCodecRoundTrip(t *testing.T) {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	for _, encoding := range []http.Encoding{http.BinaryV1, http.StructuredV1} {

		testCases := map[string]struct {
			codec   *http.Codec
			event   event.Event
			want    event.Event
			wantErr error
		}{
			"simple data v0.3": {
				codec: &http.Codec{Encoding: encoding},
				event: event.Event{
					Context: event.EventContextV03{
						Type:   "com.example.test",
						Source: *source,
						ID:     "ABC-123",
					}.AsV03(),
					Data: map[string]string{
						"a": "apple",
						"b": "banana",
					},
				},
				want: event.Event{
					Context: event.EventContextV03{
						SpecVersion: event.CloudEventsVersionV03,
						Type:        "com.example.test",
						Source:      *source,
						ID:          "ABC-123",
					}.AsV03(),
					Data: map[string]interface{}{
						"a": "apple",
						"b": "banana",
					},
					DataEncoded: true,
				},
			},
			"struct data v0.3": {
				codec: &http.Codec{Encoding: encoding},
				event: event.Event{
					Context: event.EventContextV03{
						Type:   "com.example.test",
						Source: *source,
						ID:     "ABC-123",
					}.AsV03(),
					Data: DataExample{
						AnInt:   42,
						AString: "testing",
					},
				},
				want: event.Event{
					Context: event.EventContextV03{
						SpecVersion: event.CloudEventsVersionV03,
						Type:        "com.example.test",
						Source:      *source,
						ID:          "ABC-123",
					}.AsV03(),
					Data: &DataExample{
						AnInt:   42,
						AString: "testing",
					},
					DataEncoded: true,
				},
			},
			// TODO: add tests for other versions. (note not really needed because these is tested internally too)
		}
		for n, tc := range testCases {
			n = fmt.Sprintf("%s, %s", encoding, n)
			t.Run(n, func(t *testing.T) {

				msg, err := tc.codec.Encode(context.TODO(), tc.event)
				if err != nil {
					if diff := cmp.Diff(tc.wantErr, err); diff != "" {
						t.Errorf("unexpected error (-want, +got) = %v", diff)
					}
					return
				}

				got, err := tc.codec.Decode(context.TODO(), msg)
				if err != nil {
					if diff := cmp.Diff(tc.wantErr, err); diff != "" {
						t.Errorf("unexpected error (-want, +got) = %v", diff)
					}
					return
				}

				if tc.event.Data != nil {
					// We have to be pretty tricky in the test to get the correct type.
					data, _ := types.Allocate(tc.want.Data)
					err := got.DataAs(&data)
					if err != nil {
						if diff := cmp.Diff(tc.wantErr, err); diff != "" {
							t.Errorf("unexpected error (-want, +got) = %v", diff)
						}
						return
					}
					got.Data = data
				}

				if tc.wantErr != nil {
					if diff := cmp.Diff(tc.wantErr, err); diff != "" {
						t.Errorf("unexpected error (-want, +got) = %v", diff)
					}
					return
				}

				// fix the context back to v03 to test.
				ctxv03 := got.Context.AsV03()
				got.Context = ctxv03

				if diff := cmp.Diff(tc.want, *got); diff != "" {
					t.Errorf("unexpected event (-want, +got) = %v", diff)
				}
			})
		}

	}
}

func toBytes(body map[string]interface{}) []byte {
	b, err := json.Marshal(body)
	if err != nil {
		return []byte(fmt.Sprintf(`{"error":%q}`, err.Error()))
	}
	return b
}
