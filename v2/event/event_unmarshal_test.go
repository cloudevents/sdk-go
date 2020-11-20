package event_test

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

func TestUnmarshal(t *testing.T) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URIRef{URL: *sourceUrl}
	sourceV1 := &types.URIRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URIRef{URL: *schemaUrl}
	schemaV1 := &types.URI{URL: *schemaUrl}

	testCases := map[string]struct {
		body    []byte
		want    *event.Event
		wantErr error
	}{
		"struct data v0.3": {
			body: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "0.3",
				"datacontenttype": "application/json",
				"data": map[string]interface{}{
					"a": 42,
					"b": "testing",
				},
				"id":        "ABC-123",
				"time":      now.Format(time.RFC3339Nano),
				"type":      "com.example.test",
				"exbool":    true,
				"exint":     42,
				"exstring":  "exstring",
				"exbinary":  "AAECAw==",
				"exurl":     "http://example.com/source",
				"extime":    now.Format(time.RFC3339Nano),
				"schemaurl": "http://example.com/schema",
				"source":    "http://example.com/source",
			}),
			want: &event.Event{
				Context: event.EventContextV03{
					Type:            "com.example.test",
					Source:          *source,
					SchemaURL:       schema,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"exbool":   true, // Boolean should be preserved
						"exint":    int32(42),
						"exstring": "exstring",
						// Since byte, url and time are encoded as string, the unmarshal should just convert them to string
						"exbinary": "AAECAw==",
						"exurl":    "http://example.com/source",
						"extime":   now.Format(time.RFC3339Nano),
					},
				}.AsV03(),
				DataEncoded: mustJsonMarshal(t, DataExample{
					AnInt:   42,
					AString: "testing",
				}),
			},
		},
		"string data v0.3": {
			body: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "0.3",
				"datacontenttype": "application/json",
				"data":            "This is a string.",
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"exbool":          true,
				"exint":           42,
				"exstring":        "exstring",
				"exbinary":        "AAECAw==",
				"exurl":           "http://example.com/source",
				"extime":          now.Format(time.RFC3339Nano),
				"schemaurl":       "http://example.com/schema",
				"source":          "http://example.com/source",
			}),
			want: &event.Event{
				Context: event.EventContextV03{
					Type:            "com.example.test",
					Source:          *source,
					SchemaURL:       schema,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"exbool":   true, // Boolean should be preserved
						"exint":    int32(42),
						"exstring": "exstring",
						// Since byte, url and time are encoded as string, the unmarshal should just convert them to string
						"exbinary": "AAECAw==",
						"exurl":    "http://example.com/source",
						"extime":   now.Format(time.RFC3339Nano),
					},
				}.AsV03(),
				DataEncoded: mustJsonMarshal(t, "This is a string."),
			},
		},
		"nil data v0.3": {
			body: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "0.3",
				"datacontenttype": "application/json",
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"exbool":          true,
				"exint":           42,
				"exstring":        "exstring",
				"exbinary":        "AAECAw==",
				"exurl":           "http://example.com/source",
				"extime":          now.Format(time.RFC3339Nano),
				"schemaurl":       "http://example.com/schema",
				"source":          "http://example.com/source",
			}),
			want: &event.Event{
				Context: event.EventContextV03{
					Type:            "com.example.test",
					Source:          *source,
					SchemaURL:       schema,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"exbool":   true, // Boolean should be preserved
						"exint":    int32(42),
						"exstring": "exstring",
						// Since byte, url and time are encoded as string, the unmarshal should just convert them to string
						"exbinary": "AAECAw==",
						"exurl":    "http://example.com/source",
						"extime":   now.Format(time.RFC3339Nano),
					},
				}.AsV03(),
			},
		},
		"struct data v1.0": {
			body: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "1.0",
				"datacontenttype": "application/json",
				"data": map[string]interface{}{
					"a": 42,
					"b": "testing",
				},
				"id":         "ABC-123",
				"time":       now.Format(time.RFC3339Nano),
				"type":       "com.example.test",
				"exbool":     true,
				"exint":      42,
				"exstring":   "exstring",
				"exbinary":   "AAECAw==",
				"exurl":      "http://example.com/source",
				"extime":     now.Format(time.RFC3339Nano),
				"dataschema": "http://example.com/schema",
				"source":     "http://example.com/source",
			}),
			want: &event.Event{
				Context: event.EventContextV1{
					Type:            "com.example.test",
					Source:          *sourceV1,
					DataSchema:      schemaV1,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"exbool":   true, // Boolean should be preserved
						"exint":    int32(42),
						"exstring": "exstring",
						// Since byte, url and time are encoded as string, the unmarshal should just convert them to string
						"exbinary": "AAECAw==",
						"exurl":    "http://example.com/source",
						"extime":   now.Format(time.RFC3339Nano),
					},
				}.AsV1(),
				DataEncoded: mustJsonMarshal(t, DataExample{
					AnInt:   42,
					AString: "testing",
				}),
			},
		},
		"string data v1.0": {
			body: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "1.0",
				"datacontenttype": "application/json",
				"data":            "This is a string.",
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"exbool":          true,
				"exint":           42,
				"exstring":        "exstring",
				"exbinary":        "AAECAw==",
				"exurl":           "http://example.com/source",
				"extime":          now.Format(time.RFC3339Nano),
				"dataschema":      "http://example.com/schema",
				"source":          "http://example.com/source",
			}),
			want: &event.Event{
				Context: event.EventContextV1{
					Type:            "com.example.test",
					Source:          *sourceV1,
					DataSchema:      schemaV1,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"exbool":   true, // Boolean should be preserved
						"exint":    int32(42),
						"exstring": "exstring",
						// Since byte, url and time are encoded as string, the unmarshal should just convert them to string
						"exbinary": "AAECAw==",
						"exurl":    "http://example.com/source",
						"extime":   now.Format(time.RFC3339Nano),
					},
				}.AsV1(),
				DataEncoded: mustJsonMarshal(t, "This is a string."),
			},
		},
		"base64 json encoded data v1.0": {
			body: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "1.0",
				"datacontenttype": "application/json",
				"data_base64":     base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`)),
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"dataschema":      "http://example.com/schema",
				"source":          "http://example.com/source",
			}),
			want: &event.Event{
				Context: event.EventContextV1{
					Type:            "com.example.test",
					Source:          *sourceV1,
					DataSchema:      schemaV1,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: mustJsonMarshal(t, map[string]interface{}{"hello": "world"}),
				DataBase64:  true,
			},
		},
		"base64 xml encoded data v1.0": {
			body: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "1.0",
				"datacontenttype": "application/json",
				"data_base64":     base64.StdEncoding.EncodeToString(mustEncodeWithDataCodec(t, event.ApplicationXML, &XMLDataExample{AnInt: 10})),
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"dataschema":      "http://example.com/schema",
				"source":          "http://example.com/source",
			}),
			want: &event.Event{
				Context: event.EventContextV1{
					Type:            "com.example.test",
					Source:          *sourceV1,
					DataSchema:      schemaV1,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: mustEncodeWithDataCodec(t, event.ApplicationXML, &XMLDataExample{AnInt: 10}),
				DataBase64:  true,
			},
		},
		"xml data v1.0": {
			body: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "1.0",
				"datacontenttype": event.ApplicationXML,
				"data":            string(mustEncodeWithDataCodec(t, event.ApplicationXML, XMLDataExample{AnInt: 5, AString: "aaa"})),
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"dataschema":      "http://example.com/schema",
				"source":          "http://example.com/source",
			}),
			want: &event.Event{
				Context: event.EventContextV1{
					Type:            "com.example.test",
					Source:          *sourceV1,
					DataSchema:      schemaV1,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationXML(),
				}.AsV1(),
				DataBase64:  false,
				DataEncoded: mustEncodeWithDataCodec(t, event.ApplicationXML, &XMLDataExample{AnInt: 5, AString: "aaa"}),
			},
		},
		"nil data v1.0": {
			body: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "1.0",
				"datacontenttype": "application/json",
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"exbool":          true,
				"exint":           42,
				"exstring":        "exstring",
				"exbinary":        "AAECAw==",
				"exurl":           "http://example.com/source",
				"extime":          now.Format(time.RFC3339Nano),
				"dataschema":      "http://example.com/schema",
				"source":          "http://example.com/source",
			}),
			want: &event.Event{
				Context: event.EventContextV1{
					Type:            "com.example.test",
					Source:          *sourceV1,
					DataSchema:      schemaV1,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"exbool":   true, // Boolean should be preserved
						"exint":    int32(42),
						"exstring": "exstring",
						// Since byte, url and time are encoded as string, the unmarshal should just convert them to string
						"exbinary": "AAECAw==",
						"exurl":    "http://example.com/source",
						"extime":   now.Format(time.RFC3339Nano),
					},
				}.AsV1(),
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := &event.Event{}
			err := json.Unmarshal(tc.body, got)

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
