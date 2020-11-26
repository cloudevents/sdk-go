package event_test

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

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
		"array data v1.0": {
			body: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "1.0",
				"datacontenttype": "application/json",
				"data":            []string{"This is a string array"},
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
				DataEncoded: mustJsonMarshal(t, []string{"This is a string array"}),
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

func TestUnmarshalWithOrdering(t *testing.T) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	sourceV1 := &types.URIRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schemaV1 := &types.URI{URL: *schemaUrl}
	schemaV03 := &types.URIRef{URL: *schemaUrl}

	structData := DataExample{
		AnInt:   42,
		AString: "testing",
	}

	testCases := map[string]struct {
		body string
		want *event.Event
	}{
		"base64 json encoded data v1.0 with specversion -> datacontenttype -> data_base64": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("specversion", "1.0").
				Add("datacontenttype", "application/json").
				Add("data_base64", base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`))).
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("dataschema", "http://example.com/schema").
				Add("source", "http://example.com/source").
				End(),
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
		"base64 json encoded data v1.0 with specversion -> data_base64 -> datacontenttype": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("specversion", "1.0").
				Add("data_base64", base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`))).
				Add("datacontenttype", "application/json").
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("dataschema", "http://example.com/schema").
				Add("source", "http://example.com/source").
				End(),
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
		"base64 json encoded data v1.0 with data_base64 -> specversion -> datacontenttype": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("dataschema", "http://example.com/schema").
				Add("source", "http://example.com/source").
				Add("data_base64", base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`))).
				Add("specversion", "1.0").
				Add("datacontenttype", "application/json").
				End(),
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
		"base64 json encoded data v1.0 with data_base64 -> datacontenttype -> specversion": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("dataschema", "http://example.com/schema").
				Add("source", "http://example.com/source").
				Add("data_base64", base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`))).
				Add("datacontenttype", "application/json").
				Add("specversion", "1.0").
				End(),
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
		"struct json data v1.0 and datacontentencoding ext with data -> datacontenttype -> specversion": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("dataschema", "http://example.com/schema").
				Add("source", "http://example.com/source").
				Add("data", structData).
				Add("datacontenttype", "application/json").
				Add("datacontentencoding", "base64").
				Add("specversion", "1.0").
				End(),
			want: &event.Event{
				Context: event.EventContextV1{
					Type:            "com.example.test",
					Source:          *sourceV1,
					DataSchema:      schemaV1,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"datacontentencoding": "base64",
					},
				}.AsV1(),
				DataEncoded: mustJsonMarshal(t, structData),
				DataBase64:  false,
			},
		},
		"struct json data v1.0 with specversion -> datacontenttype -> data": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("specversion", "1.0").
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("dataschema", "http://example.com/schema").
				Add("source", "http://example.com/source").
				Add("datacontenttype", "application/json").
				Add("data", structData).
				End(),
			want: &event.Event{
				Context: event.EventContextV1{
					Type:            "com.example.test",
					Source:          *sourceV1,
					DataSchema:      schemaV1,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
				}.AsV1(),
				DataEncoded: mustJsonMarshal(t, structData),
				DataBase64:  false,
			},
		},
		"more than 16 attributes with struct data and specversion as last attribute": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("dataschema", "http://example.com/schema").
				Add("source", "http://example.com/source").
				Add("data", structData).
				Add("subject", "sub").
				Add("datacontenttype", "application/json").
				Add("ext1", "ext1").
				Add("ext2", "ext2").
				Add("ext3", "ext3").
				Add("ext4", "ext4").
				Add("ext5", "ext5").
				Add("ext6", "ext6").
				Add("ext7", "ext7").
				Add("ext8", "ext8").
				Add("ext9", "ext9").
				Add("ext10", "ext10").
				Add("specversion", "1.0").
				End(),
			want: &event.Event{
				Context: event.EventContextV1{
					Type:            "com.example.test",
					Source:          *sourceV1,
					DataSchema:      schemaV1,
					ID:              "ABC-123",
					Subject:         strptr("sub"),
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"ext1":  "ext1",
						"ext2":  "ext2",
						"ext3":  "ext3",
						"ext4":  "ext4",
						"ext5":  "ext5",
						"ext6":  "ext6",
						"ext7":  "ext7",
						"ext8":  "ext8",
						"ext9":  "ext9",
						"ext10": "ext10",
					},
				}.AsV1(),
				DataEncoded: mustJsonMarshal(t, structData),
				DataBase64:  false,
			},
		},
		"more than 16 attributes with struct data and specversion as first attribute v1.0": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("specversion", "1.0").
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("dataschema", "http://example.com/schema").
				Add("source", "http://example.com/source").
				Add("datacontenttype", "application/json").
				Add("data", structData).
				Add("subject", "sub").
				Add("ext1", "ext1").
				Add("ext2", "ext2").
				Add("ext3", "ext3").
				Add("ext4", "ext4").
				Add("ext5", "ext5").
				Add("ext6", "ext6").
				Add("ext7", "ext7").
				Add("ext8", "ext8").
				Add("ext9", "ext9").
				Add("ext10", "ext10").
				End(),
			want: &event.Event{
				Context: event.EventContextV1{
					Type:            "com.example.test",
					Source:          *sourceV1,
					DataSchema:      schemaV1,
					ID:              "ABC-123",
					Subject:         strptr("sub"),
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"ext1":  "ext1",
						"ext2":  "ext2",
						"ext3":  "ext3",
						"ext4":  "ext4",
						"ext5":  "ext5",
						"ext6":  "ext6",
						"ext7":  "ext7",
						"ext8":  "ext8",
						"ext9":  "ext9",
						"ext10": "ext10",
					},
				}.AsV1(),
				DataEncoded: mustJsonMarshal(t, structData),
				DataBase64:  false,
			},
		},
		"more than 16 attributes with struct data and specversion as first attribute v0.3": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("specversion", "0.3").
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("schemaurl", "http://example.com/schema").
				Add("source", "http://example.com/source").
				Add("datacontenttype", "application/json").
				Add("data", structData).
				Add("subject", "sub").
				Add("ext1", "ext1").
				Add("ext2", "ext2").
				Add("ext3", "ext3").
				Add("ext4", "ext4").
				Add("ext5", "ext5").
				Add("ext6", "ext6").
				Add("ext7", "ext7").
				Add("ext8", "ext8").
				Add("ext9", "ext9").
				Add("ext10", "ext10").
				End(),
			want: &event.Event{
				Context: event.EventContextV03{
					Type:            "com.example.test",
					Source:          *sourceV1,
					SchemaURL:       schemaV03,
					ID:              "ABC-123",
					Subject:         strptr("sub"),
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"ext1":  "ext1",
						"ext2":  "ext2",
						"ext3":  "ext3",
						"ext4":  "ext4",
						"ext5":  "ext5",
						"ext6":  "ext6",
						"ext7":  "ext7",
						"ext8":  "ext8",
						"ext9":  "ext9",
						"ext10": "ext10",
					},
				}.AsV03(),
				DataEncoded: mustJsonMarshal(t, structData),
				DataBase64:  false,
			},
		},
		"base64 json encoded data v0.3 with specversion -> datacontenttype -> datacontentencoding -> data": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("specversion", "0.3").
				Add("datacontenttype", "application/json").
				Add("datacontentencoding", "base64").
				Add("data", base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`))).
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("source", "http://example.com/source").
				End(),
			want: &event.Event{
				Context: event.EventContextV03{
					Type:                "com.example.test",
					Source:              *sourceV1,
					ID:                  "ABC-123",
					Time:                &now,
					DataContentEncoding: strptr(event.Base64),
					DataContentType:     event.StringOfApplicationJSON(),
				}.AsV03(),
				DataEncoded: mustJsonMarshal(t, map[string]interface{}{"hello": "world"}),
				DataBase64:  true,
			},
		},
		"base64 json encoded data v0.3 with specversion -> datacontenttype -> data -> datacontentencoding": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("specversion", "0.3").
				Add("datacontenttype", "application/json").
				Add("data", base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`))).
				Add("datacontentencoding", "base64").
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("source", "http://example.com/source").
				End(),
			want: &event.Event{
				Context: event.EventContextV03{
					Type:                "com.example.test",
					Source:              *sourceV1,
					ID:                  "ABC-123",
					Time:                &now,
					DataContentEncoding: strptr(event.Base64),
					DataContentType:     event.StringOfApplicationJSON(),
				}.AsV03(),
				DataEncoded: mustJsonMarshal(t, map[string]interface{}{"hello": "world"}),
				DataBase64:  true,
			},
		},
		"base64 json encoded data v0.3 with datacontenttype -> data -> specversion -> datacontentencoding": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("source", "http://example.com/source").
				Add("datacontenttype", "application/json").
				Add("data", base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`))).
				Add("specversion", "0.3").
				Add("datacontentencoding", "base64").
				End(),
			want: &event.Event{
				Context: event.EventContextV03{
					Type:                "com.example.test",
					Source:              *sourceV1,
					ID:                  "ABC-123",
					Time:                &now,
					DataContentEncoding: strptr(event.Base64),
					DataContentType:     event.StringOfApplicationJSON(),
				}.AsV03(),
				DataEncoded: mustJsonMarshal(t, map[string]interface{}{"hello": "world"}),
				DataBase64:  true,
			},
		},
		"base64 json encoded data v0.3 with datacontenttype -> data -> datacontentencoding -> specversion": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("source", "http://example.com/source").
				Add("datacontenttype", "application/json").
				Add("data", base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`))).
				Add("datacontentencoding", "base64").
				Add("subject", "sub").
				Add("specversion", "0.3").
				End(),
			want: &event.Event{
				Context: event.EventContextV03{
					Type:                "com.example.test",
					Source:              *sourceV1,
					ID:                  "ABC-123",
					Time:                &now,
					DataContentEncoding: strptr(event.Base64),
					Subject:             strptr("sub"),
					DataContentType:     event.StringOfApplicationJSON(),
				}.AsV03(),
				DataEncoded: mustJsonMarshal(t, map[string]interface{}{"hello": "world"}),
				DataBase64:  true,
			},
		},
		"base64 json encoded data v0.3 with data -> datacontenttype -> specversion -> datacontentencoding": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("data", base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`))).
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("source", "http://example.com/source").
				Add("datacontenttype", "application/json").
				Add("specversion", "0.3").
				Add("datacontentencoding", "base64").
				End(),
			want: &event.Event{
				Context: event.EventContextV03{
					Type:                "com.example.test",
					Source:              *sourceV1,
					ID:                  "ABC-123",
					Time:                &now,
					DataContentEncoding: strptr(event.Base64),
					DataContentType:     event.StringOfApplicationJSON(),
				}.AsV03(),
				DataEncoded: mustJsonMarshal(t, map[string]interface{}{"hello": "world"}),
				DataBase64:  true,
			},
		},
		"base64 json encoded data v0.3 with data/data_base64": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("data_base64", "foo").
				Add("data", base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`))).
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("source", "http://example.com/source").
				Add("datacontenttype", "application/json").
				Add("specversion", "0.3").
				Add("datacontentencoding", "base64").
				End(),
			want: &event.Event{
				Context: event.EventContextV03{
					Type:                "com.example.test",
					Source:              *sourceV1,
					ID:                  "ABC-123",
					Time:                &now,
					DataContentEncoding: strptr(event.Base64),
					DataContentType:     event.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"data_base64": "foo",
					},
				}.AsV03(),
				DataEncoded: mustJsonMarshal(t, map[string]interface{}{"hello": "world"}),
				DataBase64:  true,
			},
		},
		"struct json data v0.3 with with data/data_base64": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("data_base64", "foo").
				Add("data", structData).
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("source", "http://example.com/source").
				Add("datacontenttype", "application/json").
				Add("specversion", "0.3").
				End(),
			want: &event.Event{
				Context: event.EventContextV03{
					Type:            "com.example.test",
					Source:          *sourceV1,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
					Extensions: map[string]interface{}{
						"data_base64": "foo",
					},
				}.AsV03(),
				DataEncoded: mustJsonMarshal(t, structData),
				DataBase64:  false,
			},
		},
		"xml data v1.0": {
			body: new(orderedJsonObjectBuilder).Start().
				Add("specversion", "1.0").
				Add("datacontenttype", event.ApplicationXML).
				Add("data", string(mustEncodeWithDataCodec(t, event.ApplicationXML, XMLDataExample{AnInt: 5, AString: "aaa"}))).
				Add("id", "ABC-123").
				Add("time", now.Format(time.RFC3339Nano)).
				Add("type", "com.example.test").
				Add("dataschema", "http://example.com/schema").
				Add("source", "http://example.com/source").
				End(),
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
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := &event.Event{}
			err := json.Unmarshal([]byte(tc.body), got)

			require.NoError(t, err)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected event (-want, +got) = %v", diff)
			}
		})
	}
}

func TestUnmarshalWithOrderingError(t *testing.T) {
	testCases := map[string]string{
		"double specversion": new(orderedJsonObjectBuilder).Start().
			Add("specversion", "1.0").
			Add("specversion", "1.0.1").
			Add("id", "ABC-123").
			Add("type", "com.example.test").
			Add("dataschema", "http://example.com/schema").
			Add("source", "http://example.com/source").
			End(),
		"wrong specversion": new(orderedJsonObjectBuilder).Start().
			Add("specversion", "9000.1").
			Add("id", "ABC-123").
			Add("type", "com.example.test").
			Add("dataschema", "http://example.com/schema").
			Add("source", "http://example.com/source").
			End(),
		"no specversion": new(orderedJsonObjectBuilder).Start().
			Add("id", "ABC-123").
			Add("type", "com.example.test").
			Add("dataschema", "http://example.com/schema").
			Add("source", "http://example.com/source").
			End(),
		"wrong time": new(orderedJsonObjectBuilder).Start().
			Add("specversion", "1.0").
			Add("id", "ABC-123").
			Add("time", "afsa").
			Add("type", "com.example.test").
			Add("dataschema", "http://example.com/schema").
			Add("source", "http://example.com/source").
			End(),
		"invalid datacontentencoding with specversion -> datacontentencoding": new(orderedJsonObjectBuilder).Start().
			Add("id", "ABC-123").
			Add("type", "com.example.test").
			Add("source", "http://example.com/source").
			Add("datacontenttype", "application/json").
			Add("data", base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`))).
			Add("datacontentencoding", "base54").
			Add("specversion", "0.3").
			End(),
		"invalid datacontentencoding": new(orderedJsonObjectBuilder).Start().
			Add("specversion", "0.3").
			Add("id", "ABC-123").
			Add("type", "com.example.test").
			Add("source", "http://example.com/source").
			Add("datacontenttype", "application/json").
			Add("data", base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`))).
			Add("datacontentencoding", "base54").
			End(),
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := &event.Event{}
			err := json.Unmarshal([]byte(tc), got)

			require.Error(t, err)
		})
	}
}

// This is a little hack we need to create a json ordered.
// This makes the bench reproducible for unmarshal
type orderedJsonObjectBuilder strings.Builder

func (b *orderedJsonObjectBuilder) Start() *orderedJsonObjectBuilder {
	(*strings.Builder)(b).WriteRune('{')

	return b
}

func (b *orderedJsonObjectBuilder) Add(key string, value interface{}) *orderedJsonObjectBuilder {
	(*strings.Builder)(b).WriteRune('"')
	(*strings.Builder)(b).WriteString(key)
	(*strings.Builder)(b).WriteString("\":")

	v, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	(*strings.Builder)(b).Write(v)
	(*strings.Builder)(b).WriteRune(',')

	return b
}

func (b *orderedJsonObjectBuilder) End() string {
	str := (*strings.Builder)(b).String()
	return str[0:len(str)-1] + "}"
}
