package http_test

import (
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
	"time"
)

func strptr(s string) *string {
	return &s
}

func TestDefaultBinaryEncodingSelectionStrategy(t *testing.T) {
	testCases := map[string]struct {
		event cloudevents.Event
		want  http.Encoding
	}{
		"default, unknown version": {
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					CloudEventsVersion: "unknown",
				},
			},
			want: http.Default,
		},
		"v0.1": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV01{}.AsV01(),
			},
			want: http.BinaryV01,
		},
		"v0.2": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV02{}.AsV02(),
			},
			want: http.BinaryV02,
		},
		"v0.3": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV03{}.AsV03(),
			},
			want: http.BinaryV03,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := http.DefaultBinaryEncodingSelectionStrategy(tc.event)

			if got != tc.want {
				t.Errorf("unexpected selection want: %s, got: %s", tc.want, got)
			}
		})
	}
}

func TestDefaultStructuredEncodingSelectionStrategy(t *testing.T) {
	testCases := map[string]struct {
		event cloudevents.Event
		want  http.Encoding
	}{
		"default, unknown version": {
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					CloudEventsVersion: "unknown",
				},
			},
			want: http.Default,
		},
		"v0.1": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV01{}.AsV01(),
			},
			want: http.StructuredV01,
		},
		"v0.2": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV02{}.AsV02(),
			},
			want: http.StructuredV02,
		},
		"v0.3": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV03{}.AsV03(),
			},
			want: http.StructuredV03,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := http.DefaultStructuredEncodingSelectionStrategy(tc.event)

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
		codec   http.Codec
		event   cloudevents.Event
		want    *http.Message
		wantErr error
	}{
		"default v0.1 binary": {
			codec: http.Codec{
				DefaultEncodingSelectionFn: http.DefaultBinaryEncodingSelectionStrategy,
			},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					EventType: "com.example.test",
					Source:    *source,
					EventID:   "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"CE-CloudEventsVersion": {"0.1"},
					"CE-EventID":            {"ABC-123"},
					"CE-EventType":          {"com.example.test"},
					"CE-Source":             {"http://example.com/source"},
					"Content-Type":          {"application/json"},
				},
			},
		},
		"default v0.1 structured": {
			codec: http.Codec{
				DefaultEncodingSelectionFn: http.DefaultStructuredEncodingSelectionStrategy,
			},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					EventType: "com.example.test",
					Source:    *source,
					EventID:   "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"contentType":        "application/json",
						"cloudEventsVersion": "0.1",
						"eventID":            "ABC-123",
						"eventType":          "com.example.test",
						"source":             "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"default v0.2 binary": {
			codec: http.Codec{
				DefaultEncodingSelectionFn: http.DefaultBinaryEncodingSelectionStrategy,
			},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
		},
		"default v0.2 structured": {
			codec: http.Codec{
				DefaultEncodingSelectionFn: http.DefaultStructuredEncodingSelectionStrategy,
			},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"contenttype": "application/json",
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
			codec: http.Codec{
				DefaultEncodingSelectionFn: http.DefaultBinaryEncodingSelectionStrategy,
			},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.3"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
		},
		"default v0.3 structured": {
			codec: http.Codec{
				DefaultEncodingSelectionFn: http.DefaultStructuredEncodingSelectionStrategy,
			},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"datacontenttype": "application/json",
						"specversion":     "0.3",
						"id":              "ABC-123",
						"type":            "com.example.test",
						"source":          "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"simple v0.1 binary": {
			codec: http.Codec{Encoding: http.BinaryV01},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					EventType: "com.example.test",
					Source:    *source,
					EventID:   "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"CE-CloudEventsVersion": {"0.1"},
					"CE-EventID":            {"ABC-123"},
					"CE-EventType":          {"com.example.test"},
					"CE-Source":             {"http://example.com/source"},
					"Content-Type":          {"application/json"},
				},
			},
		},
		"simple v0.1 structured": {
			codec: http.Codec{Encoding: http.StructuredV01},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					EventType: "com.example.test",
					Source:    *source,
					EventID:   "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"contentType":        "application/json",
						"cloudEventsVersion": "0.1",
						"eventID":            "ABC-123",
						"eventType":          "com.example.test",
						"source":             "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"simple v0.2 binary": {
			codec: http.Codec{Encoding: http.BinaryV02},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
		},
		"simple v0.2 structured": {
			codec: http.Codec{Encoding: http.StructuredV02},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"contenttype": "application/json",
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
			codec: http.Codec{Encoding: http.BinaryV03},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.3"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
		},
		"simple v0.3 structured": {
			codec: http.Codec{Encoding: http.StructuredV03},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"datacontenttype": "application/json",
						"specversion":     "0.3",
						"id":              "ABC-123",
						"type":            "com.example.test",
						"source":          "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, err := tc.codec.Encode(tc.event)

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

func TestCodecDecode(t *testing.T) {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	testCases := map[string]struct {
		codec   http.Codec
		msg     *http.Message
		want    *cloudevents.Event
		wantErr error
	}{
		"simple v0.1 binary": {
			codec: http.Codec{Encoding: http.BinaryV01},
			msg: &http.Message{
				Header: map[string][]string{
					"CE-CloudEventsVersion": {"0.1"},
					"CE-EventID":            {"ABC-123"},
					"CE-EventType":          {"com.example.test"},
					"CE-Source":             {"http://example.com/source"},
					"Content-Type":          {"application/json"},
				},
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					CloudEventsVersion: cloudevents.CloudEventsVersionV01,
					EventType:          "com.example.test",
					Source:             *source,
					EventID:            "ABC-123",
					ContentType:        cloudevents.StringOfApplicationJSON(),
				},
				DataEncoded: true,
			},
		},
		"simple v0.1 structured": {
			codec: http.Codec{Encoding: http.StructuredV01},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"cloudEventsVersion": "0.1",
						"eventID":            "ABC-123",
						"eventType":          "com.example.test",
						"source":             "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					CloudEventsVersion: cloudevents.CloudEventsVersionV01,
					EventType:          "com.example.test",
					Source:             *source,
					EventID:            "ABC-123",
				},
				DataEncoded: true,
			},
		},
		"simple v0.2 binary": {
			codec: http.Codec{Encoding: http.BinaryV02},
			msg: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					SpecVersion: cloudevents.CloudEventsVersionV02,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
					ContentType: cloudevents.StringOfApplicationJSON(),
				},
				DataEncoded: true,
			},
		},
		"simple v0.2 structured": {
			codec: http.Codec{Encoding: http.StructuredV02},
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
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					SpecVersion: cloudevents.CloudEventsVersionV02,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
				},
				DataEncoded: true,
			},
		},

		"simple v0.3 binary": {
			codec: http.Codec{Encoding: http.BinaryV03},
			msg: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.3"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					SpecVersion:     cloudevents.CloudEventsVersionV03,
					Type:            "com.example.test",
					Source:          *source,
					ID:              "ABC-123",
					DataContentType: cloudevents.StringOfApplicationJSON(),
				},
				DataEncoded: true,
			},
		},
		"simple v0.3 structured": {
			codec: http.Codec{Encoding: http.StructuredV03},
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
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					SpecVersion: cloudevents.CloudEventsVersionV03,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
				},
				DataEncoded: true,
			},
		},

		// Conversion tests.

		"simple v0.1 binary -> v0.2 binary": {
			codec: http.Codec{Encoding: http.BinaryV02},
			msg: &http.Message{
				Header: map[string][]string{
					"CE-CloudEventsVersion": {"0.1"},
					"CE-EventID":            {"ABC-123"},
					"CE-EventType":          {"com.example.test"},
					"CE-Source":             {"http://example.com/source"},
					"Content-Type":          {"application/json"},
				},
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					SpecVersion: cloudevents.CloudEventsVersionV02,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
					ContentType: cloudevents.StringOfApplicationJSON(),
				},
				DataEncoded: true,
			},
		},
		"simple v0.1 structured -> v0.2 structured": {
			codec: http.Codec{Encoding: http.StructuredV02},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"cloudEventsVersion": "0.1",
						"eventID":            "ABC-123",
						"eventType":          "com.example.test",
						"source":             "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					SpecVersion: cloudevents.CloudEventsVersionV02,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
				},
				DataEncoded: true,
			},
		},
		"simple v0.2 binary -> v0.1 binary": {
			codec: http.Codec{Encoding: http.BinaryV01},
			msg: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					CloudEventsVersion: cloudevents.CloudEventsVersionV01,
					EventType:          "com.example.test",
					Source:             *source,
					EventID:            "ABC-123",
					ContentType:        cloudevents.StringOfApplicationJSON(),
				},
				DataEncoded: true,
			},
		},
		"simple v0.2 structured -> v0.1 structured": {
			codec: http.Codec{Encoding: http.StructuredV01},
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
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					CloudEventsVersion: cloudevents.CloudEventsVersionV01,
					EventType:          "com.example.test",
					Source:             *source,
					EventID:            "ABC-123",
				},
				DataEncoded: true,
			},
		},
		// TODO:: add the v0.3 conversion tests. Might want to think of a new way to do this.
		// The tests are getting very long...
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got, err := tc.codec.Decode(tc.msg)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(tc.wantErr, err); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected event (-want, +got) = %v", diff)
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

	for _, encoding := range []http.Encoding{http.BinaryV01, http.BinaryV02, http.StructuredV01, http.StructuredV02} {

		testCases := map[string]struct {
			codec   http.Codec
			event   cloudevents.Event
			want    cloudevents.Event
			wantErr error
		}{
			"simple data v0.1": {
				codec: http.Codec{Encoding: encoding},
				event: cloudevents.Event{
					Context: &cloudevents.EventContextV01{
						EventType: "com.example.test",
						Source:    *source,
						EventID:   "ABC-123",
					},
					Data: map[string]string{
						"a": "apple",
						"b": "banana",
					},
				},
				want: cloudevents.Event{
					Context: &cloudevents.EventContextV01{
						CloudEventsVersion: cloudevents.CloudEventsVersionV01,
						EventType:          "com.example.test",
						Source:             *source,
						EventID:            "ABC-123",
						ContentType:        cloudevents.StringOfApplicationJSON(),
					},
					Data: map[string]interface{}{
						"a": "apple",
						"b": "banana",
					},
					DataEncoded: true,
				},
			},
			"struct data v0.1": {
				codec: http.Codec{Encoding: encoding},
				event: cloudevents.Event{
					Context: &cloudevents.EventContextV01{
						EventType: "com.example.test",
						Source:    *source,
						EventID:   "ABC-123",
					},
					Data: DataExample{
						AnInt:   42,
						AString: "testing",
					},
				},
				want: cloudevents.Event{
					Context: &cloudevents.EventContextV01{
						CloudEventsVersion: cloudevents.CloudEventsVersionV01,
						EventType:          "com.example.test",
						Source:             *source,
						EventID:            "ABC-123",
						ContentType:        cloudevents.StringOfApplicationJSON(),
					},
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

				msg, err := tc.codec.Encode(tc.event)
				if err != nil {
					if diff := cmp.Diff(tc.wantErr, err); diff != "" {
						t.Errorf("unexpected error (-want, +got) = %v", diff)
					}
					return
				}

				got, err := tc.codec.Decode(msg)
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

				if tc.wantErr != nil || err != nil {
					if diff := cmp.Diff(tc.wantErr, err); diff != "" {
						t.Errorf("unexpected error (-want, +got) = %v", diff)
					}
					return
				}

				// fix the context back to v1 to test.
				ctxv1 := got.Context.AsV01()
				got.Context = ctxv1

				if diff := cmp.Diff(tc.want, *got); diff != "" {
					t.Errorf("unexpected event (-want, +got) = %v", diff)
				}
			})
		}

	}
}

// Tests v0.1 -> X -> v0.1
func TestCodecAsMiddleware(t *testing.T) {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	for _, contentType := range []string{"application/json", "application/xml"} {
		for _, encoding := range []http.Encoding{http.BinaryV01, http.BinaryV02, http.BinaryV03, http.StructuredV01, http.StructuredV02, http.StructuredV03} {

			testCases := map[string]struct {
				codec   http.Codec
				event   cloudevents.Event
				want    cloudevents.Event
				wantErr error
			}{
				// TODO: this is commented out because xml does not support maps.
				//"simple data": {
				//	codec: http.Codec{Encoding: encoding},
				//	event: cloudevents.Event{
				//		Context: &cloudevents.EventContextV01{
				//			EventType:   "com.example.test",
				//			Source:      *source,
				//			EventID:     "ABC-123",
				//			ContentType: contentType,
				//		},
				//		Data: map[string]string{
				//			"a": "apple",
				//			"b": "banana",
				//		},
				//	},
				//	want: cloudevents.Event{
				//		Context: &cloudevents.EventContextV01{
				//			CloudEventsVersion: cloudevents.CloudEventsVersionV01,
				//			EventType:          "com.example.test",
				//			Source:             *source,
				//			EventID:            "ABC-123",
				//			ContentType:        contentType,
				//		},
				//		Data: map[string]interface{}{
				//			"a": "apple",
				//			"b": "banana",
				//		},
				//	},
				//},
				"struct data": {
					codec: http.Codec{Encoding: encoding},
					event: cloudevents.Event{
						Context: &cloudevents.EventContextV01{
							EventType:   "com.example.test",
							Source:      *source,
							EventID:     "ABC-123",
							ContentType: strptr(contentType),
						},
						Data: DataExample{
							AnInt:   42,
							AString: "testing",
						},
					},
					want: cloudevents.Event{
						Context: &cloudevents.EventContextV01{
							CloudEventsVersion: cloudevents.CloudEventsVersionV01,
							EventType:          "com.example.test",
							Source:             *source,
							EventID:            "ABC-123",
							ContentType:        strptr(contentType),
						},
						Data: &DataExample{
							AnInt:   42,
							AString: "testing",
						},
						DataEncoded: true,
					},
				},
			}
			for n, tc := range testCases {
				n = fmt.Sprintf("%s[%s],%s", encoding, contentType, n)
				t.Run(n, func(t *testing.T) {

					msg1, err := tc.codec.Encode(tc.event)
					if err != nil {
						if diff := cmp.Diff(tc.wantErr, err); diff != "" {
							t.Errorf("unexpected error (-want, +got) = %v", diff)
						}
						return
					}

					midEvent, err := tc.codec.Decode(msg1)
					if err != nil {
						if diff := cmp.Diff(tc.wantErr, err); diff != "" {
							t.Errorf("unexpected error (-want, +got) = %v", diff)
						}
						return
					}

					msg2, err := tc.codec.Encode(*midEvent)
					if err != nil {
						if diff := cmp.Diff(tc.wantErr, err); diff != "" {
							t.Errorf("unexpected error (-want, +got) = %v", diff)
						}
						return
					}

					got, err := tc.codec.Decode(msg2)
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

					if tc.wantErr != nil || err != nil {
						if diff := cmp.Diff(tc.wantErr, err); diff != "" {
							t.Errorf("unexpected error (-want, +got) = %v", diff)
						}
						return
					}

					// fix the context back to v1 to test.
					ctxv1 := got.Context.AsV01()
					got.Context = ctxv1

					if diff := cmp.Diff(tc.want, *got); diff != "" {
						t.Errorf("unexpected event (-want, +got) = %v", diff)
					}
				})
			}
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
