package event_test

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

type XMLDataExample struct {
	AnInt   int    `xml:"a,omitempty"`
	AString string `xml:"b,omitempty"`
}

func TestMarshal(t *testing.T) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	source := &types.URIRef{URL: *sourceUrl}
	sourceV1 := &types.URIRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schema := &types.URIRef{URL: *schemaUrl}
	schemaV1 := &types.URI{URL: *schemaUrl}

	testCases := map[string]struct {
		event           event.Event
		eventExtensions map[string]interface{}
		want            []byte
		wantErr         *string
	}{
		"empty struct": {
			event:   event.Event{},
			wantErr: strptr("json: error calling MarshalJSON for type event.Event: specversion: missing Event.Context\n"),
		},
		"struct data v0.3": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV03{
						Type:      "com.example.test",
						Source:    *source,
						SchemaURL: schema,
						ID:        "ABC-123",
						Time:      &now,
					}.AsV03(),
				}
				_ = e.SetData(event.ApplicationJSON, DataExample{
					AnInt:   42,
					AString: "testing",
				})
				return e
			}(),
			eventExtensions: map[string]interface{}{
				"exbool":   true,
				"exint":    int32(42),
				"exstring": "exstring",
				"exbinary": []byte{0, 1, 2, 3},
				"exurl":    source,
				"extime":   &now,
			},
			want: mustJsonMarshal(t, map[string]interface{}{
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
		},
		"nil data v0.3": {
			event: event.Event{
				Context: event.EventContextV03{
					Type:            "com.example.test",
					Source:          *source,
					SchemaURL:       schema,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
				}.AsV03(),
			},
			eventExtensions: map[string]interface{}{
				"exbool":   true,
				"exint":    int32(42),
				"exstring": "exstring",
				"exbinary": []byte{0, 1, 2, 3},
				"exurl":    source,
				"extime":   &now,
			},
			want: mustJsonMarshal(t, map[string]interface{}{
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
		},
		"string data v0.3": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV03{
						Type:      "com.example.test",
						Source:    *source,
						SchemaURL: schema,
						ID:        "ABC-123",
						Time:      &now,
					}.AsV03(),
				}
				_ = e.SetData(event.ApplicationJSON, "This is a string.")
				return e
			}(),
			eventExtensions: map[string]interface{}{
				"exbool":   true,
				"exint":    int32(42),
				"exstring": "exstring",
				"exbinary": []byte{0, 1, 2, 3},
				"exurl":    source,
				"extime":   &now,
			},
			want: mustJsonMarshal(t, map[string]interface{}{
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
		},
		"struct data v1.0": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV1{
						Type:       "com.example.test",
						Source:     *sourceV1,
						DataSchema: schemaV1,
						ID:         "ABC-123",
						Time:       &now,
					}.AsV1(),
				}
				_ = e.SetData(event.ApplicationJSON, DataExample{
					AnInt:   42,
					AString: "testing",
				})
				return e
			}(),
			eventExtensions: map[string]interface{}{
				"exbool":   true,
				"exint":    int32(42),
				"exstring": "exstring",
				"exbinary": []byte{0, 1, 2, 3},
				"exurl":    sourceV1,
				"extime":   &now,
			},
			want: mustJsonMarshal(t, map[string]interface{}{
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
		},
		"nil data v1.0": {
			event: event.Event{
				Context: event.EventContextV1{
					Type:            "com.example.test",
					Source:          *sourceV1,
					DataSchema:      schemaV1,
					ID:              "ABC-123",
					Time:            &now,
					DataContentType: event.StringOfApplicationJSON(),
				}.AsV1(),
			},
			eventExtensions: map[string]interface{}{
				"exbool":   true,
				"exint":    int32(42),
				"exstring": "exstring",
				"exbinary": []byte{0, 1, 2, 3},
				"exurl":    sourceV1,
				"extime":   &now,
			},
			want: mustJsonMarshal(t, map[string]interface{}{
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
		},
		"string data v1.0": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV1{
						Type:       "com.example.test",
						Source:     *sourceV1,
						DataSchema: schemaV1,
						ID:         "ABC-123",
						Time:       &now,
					}.AsV1(),
				}
				_ = e.SetData(event.ApplicationJSON, "This is a string.")
				return e
			}(),
			eventExtensions: map[string]interface{}{
				"exbool":   true,
				"exint":    int32(42),
				"exstring": "exstring",
				"exbinary": []byte{0, 1, 2, 3},
				"exurl":    sourceV1,
				"extime":   &now,
			},
			want: mustJsonMarshal(t, map[string]interface{}{
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
		},
		"base64 json encoded data v1.0": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV1{
						Type:       "com.example.test",
						Source:     *sourceV1,
						DataSchema: schemaV1,
						ID:         "ABC-123",
						Time:       &now,
					}.AsV1(),
				}
				_ = e.SetData(event.ApplicationJSON, []byte(`{"hello": "world"}`))
				return e
			}(),
			want: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "1.0",
				"datacontenttype": "application/json",
				"data_base64":     []byte(`{"hello": "world"}`),
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"dataschema":      "http://example.com/schema",
				"source":          "http://example.com/source",
			}),
		},
		"base64 xml encoded data v1.0": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV1{
						Type:       "com.example.test",
						Source:     *sourceV1,
						DataSchema: schemaV1,
						ID:         "ABC-123",
						Time:       &now,
					}.AsV1(),
				}
				_ = e.SetData(event.ApplicationXML, mustEncodeWithDataCodec(t, event.ApplicationXML, XMLDataExample{AnInt: 5, AString: "aaa"}))
				return e
			}(),
			want: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "1.0",
				"datacontenttype": event.ApplicationXML,
				"data_base64":     mustEncodeWithDataCodec(t, event.ApplicationXML, XMLDataExample{AnInt: 5, AString: "aaa"}),
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"dataschema":      "http://example.com/schema",
				"source":          "http://example.com/source",
			}),
		},
		"xml data v1.0": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV1{
						Type:       "com.example.test",
						Source:     *sourceV1,
						DataSchema: schemaV1,
						ID:         "ABC-123",
						Time:       &now,
					}.AsV1(),
				}
				_ = e.SetData(event.ApplicationXML, XMLDataExample{AnInt: 5, AString: "aaa"})
				return e
			}(),
			want: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "1.0",
				"datacontenttype": event.ApplicationXML,
				"data":            string(mustEncodeWithDataCodec(t, event.ApplicationXML, XMLDataExample{AnInt: 5, AString: "aaa"})),
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"dataschema":      "http://example.com/schema",
				"source":          "http://example.com/source",
			}),
		},
		"number data v1.0": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV1{
						Type:       "com.example.test",
						Source:     *sourceV1,
						DataSchema: schemaV1,
						ID:         "ABC-123",
						Time:       &now,
					}.AsV1(),
				}
				_ = e.SetData(event.ApplicationJSON, 101)
				return e
			}(),
			want: mustJsonMarshal(t, map[string]interface{}{
				"specversion":     "1.0",
				"datacontenttype": "application/json",
				"data":            101,
				"id":              "ABC-123",
				"time":            now.Format(time.RFC3339Nano),
				"type":            "com.example.test",
				"dataschema":      "http://example.com/schema",
				"source":          "http://example.com/source",
			}),
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			event := tc.event

			for k, v := range tc.eventExtensions {
				event.SetExtension(k, v)
			}

			gotBytes, err := json.Marshal(event)

			if tc.wantErr != nil || err != nil {
				if diff := cmp.Diff(*tc.wantErr, err.Error()); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			// so we can understand the diff, turn bytes to strings
			want := string(tc.want)
			got := string(gotBytes)

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("unexpected event (-want, +got) = %v", diff)
			}
		})
	}
}

func mustJsonMarshal(tb testing.TB, body interface{}) []byte {
	b, err := json.Marshal(body)
	require.NoError(tb, err)
	return b
}
