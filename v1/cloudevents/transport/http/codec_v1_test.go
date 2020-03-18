package http_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
)

func TestCodecV1_Encode(t *testing.T) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URIRef{URL: *sourceUrl}

	subject := "resource"

	DataSchema, _ := url.Parse("http://example.com/schema")
	schema := &types.URI{URL: *DataSchema}

	testCases := map[string]struct {
		codec   http.CodecV1
		event   cloudevents.Event
		want    *http.Message
		wantErr error
	}{
		"simple v1.0 default": {
			codec: http.CodecV1{},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV1{
					Type:   "com.example.test",
					Source: *source,
					ID:     "ABC-123",
				},
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
		"full v1.0 default": {
			codec: http.CodecV1{},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV1{
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					DataSchema:      schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Subject:         &subject,
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
					"Ce-Specversion": {"1.0"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Time":        {now.Format(time.RFC3339Nano)},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Ce-Subject":     {"resource"},
					"Ce-Dataschema":  {"http://example.com/schema"},
					"Ce-Test":        {"extended"},
					"Content-Type":   {"application/json"},
				},
				Body: []byte(`{"hello":"world"}`),
			},
		},
		"simple v1.0 binary": {
			codec: http.CodecV1{DefaultEncoding: http.BinaryV1},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV1{
					Type:   "com.example.test",
					Source: *source,
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
		"full v1.0 binary": {
			codec: http.CodecV1{DefaultEncoding: http.BinaryV1},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV1{
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					DataSchema:      schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Subject:         &subject,
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
					"Ce-Specversion": {"1.0"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Time":        {now.Format(time.RFC3339Nano)},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Ce-Subject":     {"resource"},
					"Ce-Dataschema":  {"http://example.com/schema"},
					"Ce-Test":        {"extended"},
					"Content-Type":   {"application/json"},
				},
				Body: []byte(`{"hello":"world"}`),
			},
		},
		"simple v1.0 structured": {
			codec: http.CodecV1{DefaultEncoding: http.StructuredV1},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV1{
					Type:   "com.example.test",
					Source: *source,
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
		"full v1.0 structured": {
			codec: http.CodecV1{DefaultEncoding: http.StructuredV1},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					DataSchema:      schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Subject:         &subject,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				}.AsV1(),
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
						"specversion":     "1.0",
						"datacontenttype": "application/json",
						"data": map[string]interface{}{
							"hello": "world",
						},
						"id":         "ABC-123",
						"time":       now,
						"type":       "com.example.test",
						"test":       "extended",
						"dataschema": "http://example.com/schema",
						"source":     "http://example.com/source",
						"subject":    "resource",
					}
					return toBytes(body)
				}(),
			},
		},
		"full v1.0 structured base64": {
			codec: http.CodecV1{DefaultEncoding: http.StructuredV1},
			event: cloudevents.Event{
				Context: cloudevents.EventContextV1{
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					DataSchema:      schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Subject:         &subject,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				}.AsV1(),
				Data:        []byte(`{"hello":"world"}`),
				DataBinary:  true,
				DataEncoded: true,
			},
			want: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion":     "1.0",
						"datacontenttype": "application/json",
						"data_base64":     "eyJoZWxsbyI6IndvcmxkIn0=",
						"id":              "ABC-123",
						"time":            now,
						"type":            "com.example.test",
						"test":            "extended",
						"dataschema":      "http://example.com/schema",
						"source":          "http://example.com/source",
						"subject":         "resource",
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

func TestCodecV1_Decode(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URIRef{URL: *sourceUrl}

	subject := "resource"

	DataSchema, _ := url.Parse("http://example.com/schema")
	schema := &types.URI{URL: *DataSchema}

	testCases := map[string]struct {
		codec   http.CodecV1
		msg     *http.Message
		want    *cloudevents.Event
		wantErr error
	}{
		"simple v1.0 binary": {
			codec: http.CodecV1{},
			msg: &http.Message{
				Header: map[string][]string{
					"ce-specversion": {"1.0"},
					"ce-id":          {"ABC-123"},
					"ce-type":        {"com.example.test"},
					"ce-source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV1{
					SpecVersion:     cloudevents.CloudEventsVersionV1,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Type:            "com.example.test",
					Source:          *source,
					ID:              "ABC-123",
				},
			},
		},
		"full v1.0 binary": {
			codec: http.CodecV1{},
			msg: &http.Message{
				Header: map[string][]string{
					"ce-specversion": {"1.0"},
					"ce-id":          {"ABC-123"},
					"ce-time":        {now.Format(time.RFC3339Nano)},
					"ce-type":        {"com.example.test"},
					"ce-source":      {"http://example.com/source"},
					"ce-subject":     {"resource"},
					"ce-dataschema":  {"http://example.com/schema"},
					"ce-test":        {"extended binary"},
					"Content-Type":   {"application/json"},
				},
				Body: toBytes(map[string]interface{}{
					"hello": "world",
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV1{
					SpecVersion:     cloudevents.CloudEventsVersionV1,
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					DataSchema:      schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Subject:         &subject,
					Extensions: map[string]interface{}{
						"test": "extended binary",
					},
				},
				Data: toBytes(map[string]interface{}{
					"hello": "world",
				}),
				DataEncoded: true,
			},
		},
		"simple v1.0 structured": {
			codec: http.CodecV1{},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: toBytes(map[string]interface{}{
					"specversion": "1.0",
					"id":          "ABC-123",
					"type":        "com.example.test",
					"source":      "http://example.com/source",
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV1{
					SpecVersion: cloudevents.CloudEventsVersionV1,
					Type:        "com.example.test",
					Source:      *source,
					ID:          "ABC-123",
				},
			},
		},
		"full v1.0 structured": {
			codec: http.CodecV1{},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: toBytes(map[string]interface{}{
					"specversion":     "1.0",
					"datacontenttype": "application/json",
					"data": map[string]interface{}{
						"hello": "world",
					},
					"id":         "ABC-123",
					"time":       now,
					"type":       "com.example.test",
					"test":       "extended",
					"dataschema": "http://example.com/schema",
					"source":     "http://example.com/source",
					"subject":    "resource",
				}),
			},
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					SpecVersion:     cloudevents.CloudEventsVersionV1,
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					DataSchema:      schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Subject:         &subject,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				}.AsV1(),
				Data: toBytes(map[string]interface{}{
					"hello": "world",
				}),
				DataEncoded: true,
			},
		},
		"full v1.0 structured base64": {
			codec: http.CodecV1{},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: toBytes(map[string]interface{}{
					"specversion":     "1.0",
					"datacontenttype": "application/json",
					"data_base64":     "eyJoZWxsbyI6IndvcmxkIn0=",
					"id":              "ABC-123",
					"time":            now,
					"type":            "com.example.test",
					"test":            "extended",
					"dataschema":      "http://example.com/schema",
					"source":          "http://example.com/source",
					"subject":         "resource",
				}),
			},
			want: &cloudevents.Event{
				Context: cloudevents.EventContextV1{
					SpecVersion:     cloudevents.CloudEventsVersionV1,
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					DataSchema:      schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Subject:         &subject,
					Extensions: map[string]interface{}{
						"test": "extended",
					},
				}.AsV1(),
				Data: []byte(`{"hello":"world"}`),
			},
		},
		"simple v1.0 binary with short header": {
			codec: http.CodecV1{},
			msg: &http.Message{
				Header: map[string][]string{
					"ce-specversion": {"1.0"},
					"ce-id":          {"ABC-123"},
					"ce-type":        {"com.example.test"},
					"ce-source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
					"X":              {"Notice how short the header's name is"},
				},
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV1{
					SpecVersion:     cloudevents.CloudEventsVersionV1,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Type:            "com.example.test",
					Source:          *source,
					ID:              "ABC-123",
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
