package http_test

import (
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
	"time"
)

func TestCodecV01_Encode(t *testing.T) {
	now := canonical.Timestamp{Time: time.Now()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &canonical.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &canonical.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   http.CodecV01
		event   canonical.Event
		want    *http.Message
		wantErr error
	}{
		"simple v1 default": {
			codec: http.CodecV01{},
			event: canonical.Event{
				Context: canonical.EventContextV01{
					CloudEventsVersion: "TestIfDefaulted",
					EventType:          "com.example.test",
					Source:             *source,
					EventID:            "ABC-123",
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
		"full v1 default": {
			codec: http.CodecV01{},
			event: canonical.Event{
				Context: canonical.EventContextV01{
					EventID:          "ABC-123",
					EventTime:        &now,
					EventType:        "com.example.full",
					EventTypeVersion: "v1alpha1",
					SchemaURL:        schema,
					ContentType:      "application/json",
					Source:           *source,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				},
				Data: map[string]interface{}{
					"hello": "world",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"CE-CloudEventsVersion": {"0.1"},
					"CE-EventID":            {"ABC-123"},
					"CE-EventTime":          {now.Format(time.RFC3339Nano)},
					"CE-EventType":          {"com.example.full"},
					"CE-EventTypeVersion":   {"v1alpha1"},
					"CE-Source":             {"http://example.com/source"},
					"CE-SchemaURL":          {"http://example.com/schema"},
					"Content-Type":          {"application/json"},
					"CE-X-Test":             {`"extended"`},
				},
				Body: []byte(`{"hello":"world"}`),
			},
		},
		"simple v1 binary": {
			codec: http.CodecV01{Encoding: http.BinaryV01},
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
		"full v1 binary": {
			codec: http.CodecV01{Encoding: http.BinaryV01},
			event: canonical.Event{
				Context: canonical.EventContextV01{
					EventID:          "ABC-123",
					EventTime:        &now,
					EventType:        "com.example.full",
					EventTypeVersion: "v1alpha1",
					SchemaURL:        schema,
					ContentType:      "application/json",
					Source:           *source,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				},
				Data: map[string]interface{}{
					"hello": "world",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"CE-CloudEventsVersion": {"0.1"},
					"CE-EventID":            {"ABC-123"},
					"CE-EventTime":          {now.Format(time.RFC3339Nano)},
					"CE-EventType":          {"com.example.full"},
					"CE-EventTypeVersion":   {"v1alpha1"},
					"CE-Source":             {"http://example.com/source"},
					"CE-SchemaURL":          {"http://example.com/schema"},
					"Content-Type":          {"application/json"},
					"CE-X-Test":             {`"extended"`},
				},
				Body: []byte(`{"hello":"world"}`),
			},
		},
		"simple v1 structured": {
			codec: http.CodecV01{Encoding: http.StructuredV01},
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
		"full v1 structured": {
			codec: http.CodecV01{Encoding: http.StructuredV01},
			event: canonical.Event{
				Context: canonical.EventContextV01{
					EventID:          "ABC-123",
					EventTime:        &now,
					EventType:        "com.example.full",
					EventTypeVersion: "v1alpha1",
					SchemaURL:        schema,
					ContentType:      "application/json",
					Source:           *source,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				},
				Data: map[string]interface{}{
					"hello": "world",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"cloudEventsVersion": "0.1",
						"contentType":        "application/json",
						"data": map[string]interface{}{
							"hello": "world",
						},
						"eventID":          "ABC-123",
						"eventTime":        now,
						"eventType":        "com.example.full",
						"eventTypeVersion": "v1alpha1",
						"extensions": map[string]interface{}{
							"test": "extended",
						},
						"schemaURL": "http://example.com/schema",
						"source":    "http://example.com/source",
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
					t.Errorf("unexpected (-want, +got) = %v", diff)
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
				t.Errorf("unexpected (-want, +got) = %v", diff)
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
