package http_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
	"github.com/google/go-cmp/cmp"
)

func TestCodecV02_Encode(t *testing.T) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   http.CodecV02
		event   cloudevents.Event
		want    *http.Message
		wantErr error
	}{
		"simple v0.2 default": {
			codec: http.CodecV02{},
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
				},
			},
		},
		"full v0.2 default": {
			codec: http.CodecV02{},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					ID:          "ABC-123",
					Time:        &now,
					Type:        "com.example.test",
					SchemaURL:   schema,
					ContentType: cloudevents.StringOfApplicationJSON(),
					Source:      *source,
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
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Time":        {now.Format(time.RFC3339Nano)},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Ce-Schemaurl":   {"http://example.com/schema"},
					"Ce-Test":        {`"extended"`},
					"Content-Type":   {"application/json"},
				},
				Body: []byte(`{"hello":"world"}`),
			},
		},
		"simple v0.2 binary": {
			codec: http.CodecV02{DefaultEncoding: http.BinaryV02},
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
				},
			},
		},
		"full v0.2 binary": {
			codec: http.CodecV02{DefaultEncoding: http.BinaryV02},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					ID:          "ABC-123",
					Time:        &now,
					Type:        "com.example.test",
					SchemaURL:   schema,
					ContentType: cloudevents.StringOfApplicationJSON(),
					Source:      *source,
					Extensions: map[string]interface{}{
						"test": "extended",
						"asmap": map[string]interface{}{
							"a": "apple",
							"b": "banana",
							"c": map[string]interface{}{
								"d": "dog",
								"e": "eel",
							},
						},
					},
				},
				Data: map[string]interface{}{
					"hello": "world",
				},
			},
			want: &http.Message{
				Header: map[string][]string{
					"Ce-Specversion": {"0.2"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Time":        {now.Format(time.RFC3339Nano)},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Ce-Schemaurl":   {"http://example.com/schema"},
					"Ce-Test":        {`"extended"`},
					"Ce-Asmap-A":     {`"apple"`},
					"Ce-Asmap-B":     {`"banana"`},
					"Ce-Asmap-C":     {`{"d":"dog","e":"eel"}`},
					"Content-Type":   {"application/json"},
				},
				Body: []byte(`{"hello":"world"}`),
			},
		},
		"simple v0.2 structured": {
			codec: http.CodecV02{DefaultEncoding: http.StructuredV02},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV02{
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
		"full v0.2 structured": {
			codec: http.CodecV02{DefaultEncoding: http.StructuredV02},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV02{
					ID:          "ABC-123",
					Time:        &now,
					Type:        "com.example.test",
					SchemaURL:   schema,
					ContentType: cloudevents.StringOfApplicationJSON(),
					Source:      *source,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				}.AsV02(),
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
						"specversion": "0.2",
						"contenttype": "application/json",
						"data": map[string]interface{}{
							"hello": "world",
						},
						"id":        "ABC-123",
						"time":      now,
						"type":      "com.example.test",
						"test":      "extended",
						"schemaurl": "http://example.com/schema",
						"source":    "http://example.com/source",
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
						t.Errorf("unexpected message body (-want, +got) = %v", diff)
						return
					}
				}

				t.Errorf("unexpected message (-want, +got) = %v", diff)
			}
		})
	}
}

// TODO: figure out extensions for v0.2

func TestCodecV02_Decode(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   http.CodecV02
		msg     *http.Message
		want    *cloudevents.Event
		wantErr error
	}{
		"simple v0.2 binary": {
			codec: http.CodecV02{},
			msg: &http.Message{
				Header: map[string][]string{
					"ce-specversion": {"0.2"},
					"ce-id":          {"ABC-123"},
					"ce-type":        {"com.example.test"},
					"ce-source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					SpecVersion: cloudevents.CloudEventsVersionV02,
					ContentType: cloudevents.StringOfApplicationJSON(),
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
				},
			},
		},
		"full v0.2 binary": {
			codec: http.CodecV02{},
			msg: &http.Message{
				Header: map[string][]string{
					"ce-specversion": {"0.2"},
					"ce-id":          {"ABC-123"},
					"ce-time":        {now.Format(time.RFC3339Nano)},
					"ce-type":        {"com.example.test"},
					"ce-source":      {"http://example.com/source"},
					"ce-schemaurl":   {"http://example.com/schema"},
					"ce-test":        {`"extended"`},
					"ce-asmap-a":     {`"apple"`},
					"ce-asmap-b":     {`"banana"`},
					"ce-asmap-c":     {`{"d":"dog","e":"eel"}`},
					"Content-Type":   {"application/json"},
				},
				Body: toBytes(map[string]interface{}{
					"hello": "world",
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					SpecVersion: cloudevents.CloudEventsVersionV02,
					ID:          "ABC-123",
					Time:        &now,
					Type:        "com.example.test",
					SchemaURL:   schema,
					ContentType: cloudevents.StringOfApplicationJSON(),
					Source:      *source,
					Extensions: map[string]interface{}{
						"test": "extended",
						"asmap": map[string]interface{}{
							"a": []string{`"apple"`},
							"b": []string{`"banana"`},
							"c": []string{`{"d":"dog","e":"eel"}`},
						},
					},
				},
				Data: toBytes(map[string]interface{}{
					"hello": "world",
				}),
				DataEncoded: true,
			},
		},
		"simple v0.2 structured": {
			codec: http.CodecV02{},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: toBytes(map[string]interface{}{
					"specversion": "0.2",
					"id":          "ABC-123",
					"type":        "com.example.test",
					"source":      "http://example.com/source",
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					SpecVersion: cloudevents.CloudEventsVersionV02,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
				},
			},
		},
		"full v0.2 structured": {
			codec: http.CodecV02{},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: toBytes(map[string]interface{}{
					"specversion": "0.2",
					"contenttype": "application/json",
					"data": map[string]interface{}{
						"hello": "world",
					},
					"id":        "ABC-123",
					"time":      now,
					"type":      "com.example.test",
					"test":      "extended",
					"schemaurl": "http://example.com/schema",
					"source":    "http://example.com/source",
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					SpecVersion: cloudevents.CloudEventsVersionV02,
					ID:          "ABC-123",
					Time:        &now,
					Type:        "com.example.test",
					SchemaURL:   schema,
					ContentType: cloudevents.StringOfApplicationJSON(),
					Source:      *source,
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
		"simple v0.2 binary with short header": {
			codec: http.CodecV02{},
			msg: &http.Message{
				Header: map[string][]string{
					"ce-specversion": {"0.2"},
					"ce-id":          {"ABC-123"},
					"ce-type":        {"com.example.test"},
					"ce-source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
					"X":              {"Notice how short the header's name is"},
				},
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV02{
					SpecVersion: cloudevents.CloudEventsVersionV02,
					ContentType: cloudevents.StringOfApplicationJSON(),
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
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
		})
	}
}
