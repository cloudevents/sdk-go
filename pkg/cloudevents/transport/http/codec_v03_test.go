package http_test

import (
	"encoding/json"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"net/url"
	"testing"
	"time"
)

func TestCodecV03_Encode(t *testing.T) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	subject := "resource"

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   http.CodecV03
		event   cloudevents.Event
		want    *http.Message
		wantErr error
	}{
		"simple v0.3 default": {
			codec: http.CodecV03{},
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
		"full v0.3 default": {
			codec: http.CodecV03{},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					SchemaURL:       schema,
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
					"Ce-Specversion": {"0.3"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Time":        {now.Format(time.RFC3339Nano)},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Ce-Subject":     {"resource"},
					"Ce-Schemaurl":   {"http://example.com/schema"},
					"Ce-Test":        {`"extended"`},
					"Content-Type":   {"application/json"},
				},
				Body: []byte(`{"hello":"world"}`),
			},
		},
		"simple v0.3 binary": {
			codec: http.CodecV03{Encoding: http.BinaryV03},
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
		"full v0.3 binary": {
			codec: http.CodecV03{Encoding: http.BinaryV03},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					SchemaURL:       schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Subject:         &subject,
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
					"Ce-Specversion": {"0.3"},
					"Ce-Id":          {"ABC-123"},
					"Ce-Time":        {now.Format(time.RFC3339Nano)},
					"Ce-Type":        {"com.example.test"},
					"Ce-Source":      {"http://example.com/source"},
					"Ce-Subject":     {"resource"},
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
		"full v0.3 binary base64": {
			codec: http.CodecV03{Encoding: http.BinaryV03},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					ID:                  "ABC-123",
					Time:                &now,
					Type:                "com.example.test",
					SchemaURL:           schema,
					DataContentType:     cloudevents.StringOfApplicationJSON(),
					DataContentEncoding: cloudevents.StringOfBase64(),
					Source:              *source,
					Subject:             &subject,
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
					"Ce-Specversion":         {"0.3"},
					"Ce-Id":                  {"ABC-123"},
					"Ce-Time":                {now.Format(time.RFC3339Nano)},
					"Ce-Type":                {"com.example.test"},
					"Ce-Source":              {"http://example.com/source"},
					"Ce-Subject":             {"resource"},
					"Ce-Schemaurl":           {"http://example.com/schema"},
					"Ce-Test":                {`"extended"`},
					"Ce-Asmap-A":             {`"apple"`},
					"Ce-Asmap-B":             {`"banana"`},
					"Ce-Asmap-C":             {`{"d":"dog","e":"eel"}`},
					"Content-Type":           {"application/json"},
					"Ce-Datacontentencoding": {"base64"},
				},
				Body: []byte("eyJoZWxsbyI6IndvcmxkIn0="),
			},
		},
		"simple v0.3 structured": {
			codec: http.CodecV03{Encoding: http.StructuredV03},
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
		"full v0.3 structured": {
			codec: http.CodecV03{Encoding: http.StructuredV03},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					SchemaURL:       schema,
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
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: func() []byte {
					body := map[string]interface{}{
						"specversion":     "0.3",
						"datacontenttype": "application/json",
						"data": map[string]interface{}{
							"hello": "world",
						},
						"id":        "ABC-123",
						"time":      now,
						"type":      "com.example.test",
						"test":      "extended",
						"schemaurl": "http://example.com/schema",
						"source":    "http://example.com/source",
						"subject":   "resource",
					}
					return toBytes(body)
				}(),
			},
		},
		"full v0.3 structured base64": {
			codec: http.CodecV03{Encoding: http.StructuredV03},
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					ID:                  "ABC-123",
					Time:                &now,
					Type:                "com.example.test",
					SchemaURL:           schema,
					DataContentType:     cloudevents.StringOfApplicationJSON(),
					DataContentEncoding: cloudevents.StringOfBase64(),
					Source:              *source,
					Subject:             &subject,
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
						"specversion":         "0.3",
						"datacontentencoding": "base64",
						"datacontenttype":     "application/json",
						"data":                "eyJoZWxsbyI6IndvcmxkIn0=",
						"id":                  "ABC-123",
						"time":                now,
						"type":                "com.example.test",
						"test":                "extended",
						"schemaurl":           "http://example.com/schema",
						"source":              "http://example.com/source",
						"subject":             "resource",
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

// TODO: figure out extensions for v0.3

func TestCodecV03_Decode(t *testing.T) {
	now := types.Timestamp{Time: time.Now()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URLRef{URL: *sourceUrl}

	subject := "resource"

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URLRef{URL: *schemaUrl}

	testCases := map[string]struct {
		codec   http.CodecV03
		msg     *http.Message
		want    *cloudevents.Event
		wantErr error
	}{
		"simple v0.3 binary": {
			codec: http.CodecV03{},
			msg: &http.Message{
				Header: map[string][]string{
					"ce-specversion": {"0.3"},
					"ce-id":          {"ABC-123"},
					"ce-type":        {"com.example.test"},
					"ce-source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
				},
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					SpecVersion:     cloudevents.CloudEventsVersionV03,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Type:            "com.example.test",
					Source:          *source,
					ID:              "ABC-123",
				},
				DataEncoded: true,
			},
		},
		"full v0.3 binary": {
			codec: http.CodecV03{},
			msg: &http.Message{
				Header: map[string][]string{
					"ce-specversion": {"0.3"},
					"ce-id":          {"ABC-123"},
					"ce-time":        {now.Format(time.RFC3339Nano)},
					"ce-type":        {"com.example.test"},
					"ce-source":      {"http://example.com/source"},
					"ce-subject":     {"resource"},
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
				Context: &cloudevents.EventContextV03{
					SpecVersion:     cloudevents.CloudEventsVersionV03,
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					SchemaURL:       schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Subject:         &subject,
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
		"full v0.3 binary base64": {
			codec: http.CodecV03{},
			msg: &http.Message{
				Header: map[string][]string{
					"ce-specversion":         {"0.3"},
					"ce-id":                  {"ABC-123"},
					"ce-time":                {now.Format(time.RFC3339Nano)},
					"ce-type":                {"com.example.test"},
					"ce-source":              {"http://example.com/source"},
					"ce-subject":             {"resource"},
					"ce-schemaurl":           {"http://example.com/schema"},
					"ce-test":                {`"extended"`},
					"ce-asmap-a":             {`"apple"`},
					"ce-asmap-b":             {`"banana"`},
					"ce-asmap-c":             {`{"d":"dog","e":"eel"}`},
					"Content-Type":           {"application/json"},
					"ce-datacontentencoding": {"base64"},
				},
				Body: []byte("eyJoZWxsbyI6IndvcmxkIn0="),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					SpecVersion:         cloudevents.CloudEventsVersionV03,
					ID:                  "ABC-123",
					Time:                &now,
					Type:                "com.example.test",
					SchemaURL:           schema,
					DataContentType:     cloudevents.StringOfApplicationJSON(),
					Source:              *source,
					Subject:             &subject,
					DataContentEncoding: cloudevents.StringOfBase64(),
					Extensions: map[string]interface{}{
						"test": "extended",
						"asmap": map[string]interface{}{
							"a": []string{`"apple"`},
							"b": []string{`"banana"`},
							"c": []string{`{"d":"dog","e":"eel"}`},
						},
					},
				},
				Data:        []byte(`eyJoZWxsbyI6IndvcmxkIn0=`),
				DataEncoded: true,
			},
		},
		"simple v0.3 structured": {
			codec: http.CodecV03{},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: toBytes(map[string]interface{}{
					"specversion": "0.3",
					"id":          "ABC-123",
					"type":        "com.example.test",
					"source":      "http://example.com/source",
				}),
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
		"full v0.3 structured": {
			codec: http.CodecV03{},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: toBytes(map[string]interface{}{
					"specversion":     "0.3",
					"datacontenttype": "application/json",
					"data": map[string]interface{}{
						"hello": "world",
					},
					"id":        "ABC-123",
					"time":      now,
					"type":      "com.example.test",
					"test":      "extended",
					"schemaurl": "http://example.com/schema",
					"source":    "http://example.com/source",
					"subject":   "resource",
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					SpecVersion:     cloudevents.CloudEventsVersionV03,
					ID:              "ABC-123",
					Time:            &now,
					Type:            "com.example.test",
					SchemaURL:       schema,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Source:          *source,
					Subject:         &subject,
					Extensions: map[string]interface{}{
						"test": json.RawMessage(`"extended"`),
					},
				},
				Data: toBytes(map[string]interface{}{
					"hello": "world",
				}),
				DataEncoded: true,
			},
		},
		"full v0.3 structured base64": {
			codec: http.CodecV03{},
			msg: &http.Message{
				Header: map[string][]string{
					"Content-Type": {"application/cloudevents+json"},
				},
				Body: toBytes(map[string]interface{}{
					"specversion":         "0.3",
					"datacontentencoding": "base64",
					"datacontenttype":     "application/json",
					"data":                "eyJoZWxsbyI6IndvcmxkIn0=",
					"id":                  "ABC-123",
					"time":                now,
					"type":                "com.example.test",
					"test":                "extended",
					"schemaurl":           "http://example.com/schema",
					"source":              "http://example.com/source",
					"subject":             "resource",
				}),
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					SpecVersion:         cloudevents.CloudEventsVersionV03,
					ID:                  "ABC-123",
					Time:                &now,
					Type:                "com.example.test",
					SchemaURL:           schema,
					DataContentType:     cloudevents.StringOfApplicationJSON(),
					DataContentEncoding: cloudevents.StringOfBase64(),
					Source:              *source,
					Subject:             &subject,
					Extensions: map[string]interface{}{
						"test": json.RawMessage(`"extended"`),
					},
				},
				Data:        []byte(`"eyJoZWxsbyI6IndvcmxkIn0="`), // TODO: structured comes in quoted. Unquote?
				DataEncoded: true,
			},
		},
		"simple v0.3 binary with short header": {
			codec: http.CodecV03{},
			msg: &http.Message{
				Header: map[string][]string{
					"ce-specversion": {"0.3"},
					"ce-id":          {"ABC-123"},
					"ce-type":        {"com.example.test"},
					"ce-source":      {"http://example.com/source"},
					"Content-Type":   {"application/json"},
					"X":              {"Notice how short the header's name is"},
				},
			},
			want: &cloudevents.Event{
				Context: &cloudevents.EventContextV03{
					SpecVersion:     cloudevents.CloudEventsVersionV03,
					DataContentType: cloudevents.StringOfApplicationJSON(),
					Type:            "com.example.test",
					Source:          *source,
					ID:              "ABC-123",
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
