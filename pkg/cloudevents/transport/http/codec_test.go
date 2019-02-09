package http_test

import (
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
)

func TestCodecEncode(t *testing.T) {
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &canonical.URLRef{URL: *sourceUrl}

	testCases := map[string]struct {
		codec   http.Codec
		event   canonical.Event
		want    *http.Message
		wantErr error
	}{
		"simple v1 binary": {
			codec: http.Codec{Encoding: http.BinaryV01},
			event: canonical.Event{
				Context: canonical.EventContextV01{
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
		"simple v1 structured": {
			codec: http.Codec{Encoding: http.StructuredV01},
			event: canonical.Event{
				Context: canonical.EventContextV01{
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
						"cloudEventsVersion": "0.1",
						"eventID":            "ABC-123",
						"eventType":          "com.example.test",
						"source":             "http://example.com/source",
					}
					return toBytes(body)
				}(),
			},
		},
		"simple v2 binary": {
			codec: http.Codec{Encoding: http.BinaryV02},
			event: canonical.Event{
				Context: canonical.EventContextV02{
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
		"simple v2 structured": {
			codec: http.Codec{Encoding: http.StructuredV02},
			event: canonical.Event{
				Context: canonical.EventContextV02{
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
						"specversion": "0.2",
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
	source := &canonical.URLRef{URL: *sourceUrl}

	testCases := map[string]struct {
		codec   http.Codec
		msg     *http.Message
		want    *canonical.Event
		wantErr error
	}{
		"simple v1 binary": {
			codec: http.Codec{},
			msg: &http.Message{
				Header: map[string][]string{
					"CE-CloudEventsVersion": {"0.1"},
					"CE-EventID":            {"ABC-123"},
					"CE-EventType":          {"com.example.test"},
					"CE-Source":             {"http://example.com/source"},
					"Content-Type":          {"application/json"},
				},
			},
			want: &canonical.Event{
				Context: canonical.EventContextV01{
					CloudEventsVersion: canonical.CloudEventsVersionV01,
					EventType:          "com.example.test",
					Source:             *source,
					EventID:            "ABC-123",
					ContentType:        "application/json",
				},
			},
		},
		//"simple v1 structured": {
		//	codec: http.Codec{Encoding: http.StructuredV01},
		//	event: canonical.Event{
		//		Context: canonical.EventContextV01{
		//			EventType: "com.example.test",
		//			Source:    *source,
		//			EventID:   "ABC-123",
		//		},
		//	},
		//	want: &http.Message{
		//		Header: map[string][]string{
		//			"Content-Type": {"application/cloudevents+json"},
		//		},
		//		Body: func() []byte {
		//			body := map[string]interface{}{
		//				"cloudEventsVersion": "0.1",
		//				"eventID":            "ABC-123",
		//				"eventType":          "com.example.test",
		//				"source":             "http://example.com/source",
		//			}
		//			return toBytes(body)
		//		}(),
		//	},
		//},
		//"simple v2 binary": {
		//	codec: http.Codec{Encoding: http.BinaryV02},
		//	event: canonical.Event{
		//		Context: canonical.EventContextV02{
		//			Type:   "com.example.test",
		//			Source: *source,
		//			ID:     "ABC-123",
		//		},
		//	},
		//	want: &http.Message{
		//		Header: map[string][]string{
		//			"Ce-Specversion": {"0.2"},
		//			"Ce-Id":          {"ABC-123"},
		//			"Ce-Type":        {"com.example.test"},
		//			"Ce-Source":      {"http://example.com/source"},
		//			"Content-Type":   {"application/json"},
		//		},
		//	},
		//},
		//"simple v2 structured": {
		//	codec: http.Codec{Encoding: http.StructuredV02},
		//	event: canonical.Event{
		//		Context: canonical.EventContextV02{
		//			Type:   "com.example.test",
		//			Source: *source,
		//			ID:     "ABC-123",
		//		},
		//	},
		//	want: &http.Message{
		//		Header: map[string][]string{
		//			"Content-Type": {"application/cloudevents+json"},
		//		},
		//		Body: func() []byte {
		//			body := map[string]interface{}{
		//				"specversion": "0.2",
		//				"id":          "ABC-123",
		//				"type":        "com.example.test",
		//				"source":      "http://example.com/source",
		//			}
		//			return toBytes(body)
		//		}(),
		//	},
		//},
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

func toBytes(body map[string]interface{}) []byte {
	b, err := json.Marshal(body)
	if err != nil {
		return []byte(fmt.Sprintf(`{"error":%q}`, err.Error()))
	}
	return b
}
