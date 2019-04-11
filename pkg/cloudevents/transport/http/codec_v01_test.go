package http_test

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
	"time"
)

func TestCodecV01_Encode(t *testing.T) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   http.CodecV01
		event   cloudevents.Event
		want    *http.Message
		wantErr error
	}{
		"simple v0.1 default": {
			codec: http.CodecV01{},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV01{
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
		"full v0.1 default": {
			codec: http.CodecV01{},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					EventID:          "ABC-123",
					EventTime:        &now,
					EventType:        "com.example.full",
					EventTypeVersion: strptr("v1alpha1"),
					SchemaURL:        schema,
					ContentType:      cloudevents.StringOfApplicationJSON(),
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
					"CE-EventTime":          {now.UTC().Format(time.RFC3339Nano)},
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
		"simple v0.1 binary": {
			codec: http.CodecV01{Encoding: http.BinaryV01},
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
		"full v0.1 binary": {
			codec: http.CodecV01{Encoding: http.BinaryV01},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					EventID:          "ABC-123",
					EventTime:        &now,
					EventType:        "com.example.full",
					EventTypeVersion: strptr("v1alpha1"),
					SchemaURL:        schema,
					ContentType:      cloudevents.StringOfApplicationJSON(),
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
					"CE-EventTime":          {now.UTC().Format(time.RFC3339Nano)},
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
		"simple v0.1 structured": {
			codec: http.CodecV01{Encoding: http.StructuredV01},
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
		"full v0.1 structured": {
			codec: http.CodecV01{Encoding: http.StructuredV01},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					EventID:          "ABC-123",
					EventTime:        &now,
					EventType:        "com.example.full",
					EventTypeVersion: strptr("v1alpha1"),
					SchemaURL:        schema,
					ContentType:      cloudevents.StringOfApplicationJSON(),
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
						t.Errorf("unexpected message body (-want, +got) = %v", diff)
						return
					}
				}
				t.Errorf("unexpected message (-want, +got) = %v", diff)
			}
		})
	}
}

func TestCodecV01_Decode(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   http.CodecV01
		msg     *http.Message
		want    *cloudevents.Event
		wantErr error
	}{
		"simple v0.1 binary": {
			codec: http.CodecV01{},
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
		"full v0.1 binary": {
			codec: http.CodecV01{},
			msg: &http.Message{
				Header: map[string][]string{
					"CE-CloudEventsVersion": {"0.1"},
					"CE-EventID":            {"ABC-123"},
					"CE-EventTime":          {now.UTC().Format(time.RFC3339Nano)},
					"CE-EventType":          {"com.example.full"},
					"CE-EventTypeVersion":   {"v1alpha1"},
					"CE-Source":             {"http://example.com/source"},
					"CE-SchemaURL":          {"http://example.com/schema"},
					"Content-Type":          {"application/json"},
					"CE-X-Test":             {`"extended"`},
				},
				Body: toBytes(map[string]interface{}{
					"hello": "world",
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					CloudEventsVersion: cloudevents.CloudEventsVersionV01,
					EventID:            "ABC-123",
					EventTime:          &now,
					EventType:          "com.example.full",
					EventTypeVersion:   strptr("v1alpha1"),
					SchemaURL:          schema,
					ContentType:        cloudevents.StringOfApplicationJSON(),
					Source:             *source,
					Extensions: map[string]interface{}{
						"Test": "extended",
					},
				},
				Data: toBytes(map[string]interface{}{
					"hello": "world",
				}),
				DataEncoded: true,
			},
		},
		"simple v0.1 structured": {
			codec: http.CodecV01{},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: toBytes(map[string]interface{}{
					"cloudEventsVersion": "0.1",
					"eventID":            "ABC-123",
					"eventType":          "com.example.test",
					"source":             "http://example.com/source",
				}),
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
		"full v0.1 structured": {
			codec: http.CodecV01{},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: toBytes(map[string]interface{}{
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
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV01{
					CloudEventsVersion: cloudevents.CloudEventsVersionV01,
					EventID:            "ABC-123",
					EventTime:          &now,
					EventType:          "com.example.full",
					EventTypeVersion:   strptr("v1alpha1"),
					SchemaURL:          schema,
					ContentType:        cloudevents.StringOfApplicationJSON(),
					Source:             *source,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				},
				Data: toBytes(map[string]interface{}{
					"hello": "world",
				}),
				DataEncoded: true,
			},
		},
		"simple v0.1 binary with short header": {
			codec: http.CodecV01{},
			msg: &http.Message{
				Header: map[string][]string{
					"CE-CloudEventsVersion": {"0.1"},
					"CE-EventID":            {"ABC-123"},
					"CE-EventType":          {"com.example.test"},
					"CE-Source":             {"http://example.com/source"},
					"Content-Type":          {"application/json"},
					"X":                     {"Notice how short the header's name is"},
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
