package http_test

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	now := time.Now()
	source, _ := url.Parse("http://example.com/source")
	schema, _ := url.Parse("http://example.com/schema")

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
					CloudEventsVersion: canonical.CloudEventsVersionV01,
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
					CloudEventsVersion: canonical.CloudEventsVersionV01,
					EventID:            "ABC-123",
					EventTime:          now,
					EventType:          "com.example.full",
					EventTypeVersion:   "v1alpha1",
					SchemaURL:          schema,
					ContentType:        "application/json",
					Source:             *source,
					Extensions: map[string]interface{}{
						"test": string("extended"),
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
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}
