package event_test

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

var Event event.Event
var Bytes []byte
var Error error

func BenchmarkMarshal(b *testing.B) {
	now := types.Timestamp{Time: time.Now().UTC()}
	sourceUrl, _ := url.Parse("http://example.com/source")
	sourceV03 := &types.URIRef{URL: *sourceUrl}
	sourceV1 := &types.URIRef{URL: *sourceUrl}

	schemaUrl, _ := url.Parse("http://example.com/schema")
	schemaV03 := &types.URIRef{URL: *schemaUrl}
	schemaV1 := &types.URI{URL: *schemaUrl}

	testCases := map[string]struct {
		event           event.Event
		eventExtensions map[string]interface{}
	}{
		"struct data v0.3": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV03{
						Type:      "com.example.test",
						Source:    *sourceV03,
						SchemaURL: schemaV03,
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
				"exurl":    sourceV03,
				"extime":   &now,
			},
		},
		"nil data v0.3": {
			event: event.Event{
				Context: event.EventContextV03{
					Type:            "com.example.test",
					Source:          *sourceV03,
					SchemaURL:       schemaV03,
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
				"exurl":    sourceV03,
				"extime":   &now,
			},
		},
		"string data v0.3": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV03{
						Type:      "com.example.test",
						Source:    *sourceV03,
						SchemaURL: schemaV03,
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
				"exurl":    sourceV03,
				"extime":   &now,
			},
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
		},
	}
	for n, tc := range testCases {
		ev := tc.event.Clone()
		for k, v := range tc.eventExtensions {
			ev.SetExtension(k, v)
		}
		b.Run(n, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Bytes, Error = json.Marshal(ev)
			}
		})
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	now := types.Timestamp{Time: time.Now().UTC()}

	testCases := map[string]string{
		"struct data v0.3": new(orderedJsonObjectBuilder).Start().
			Add("specversion", "0.3").
			Add("datacontenttype", "application/json").
			Add("data", map[string]interface{}{
				"a": 42,
				"b": "testing",
			}).
			Add("id", "ABC-123").
			Add("time", now.Format(time.RFC3339Nano)).
			Add("type", "com.example.test").
			Add("exbool", true).
			Add("exint", 42).
			Add("exstring", "exstring").
			Add("exbinary", "AAECAw==").
			Add("exurl", "http://example.com/source").
			Add("extime", now.Format(time.RFC3339Nano)).
			Add("schemaurl", "http://example.com/schema").
			Add("source", "http://example.com/source").
			End(),
		"string data v0.3": new(orderedJsonObjectBuilder).Start().
			Add("specversion", "0.3").
			Add("datacontenttype", "application/json").
			Add("data", "This is a string.").
			Add("id", "ABC-123").
			Add("time", now.Format(time.RFC3339Nano)).
			Add("type", "com.example.test").
			Add("exbool", true).
			Add("exint", 42).
			Add("exstring", "exstring").
			Add("exbinary", "AAECAw==").
			Add("exurl", "http://example.com/source").
			Add("extime", now.Format(time.RFC3339Nano)).
			Add("schemaurl", "http://example.com/schema").
			Add("source", "http://example.com/source").
			End(),
		"nil data v0.3": new(orderedJsonObjectBuilder).Start().
			Add("specversion", "0.3").
			Add("datacontenttype", "application/json").
			Add("id", "ABC-123").
			Add("time", now.Format(time.RFC3339Nano)).
			Add("type", "com.example.test").
			Add("exbool", true).
			Add("exint", 42).
			Add("exstring", "exstring").
			Add("exbinary", "AAECAw==").
			Add("exurl", "http://example.com/source").
			Add("extime", now.Format(time.RFC3339Nano)).
			Add("schemaurl", "http://example.com/schema").
			Add("source", "http://example.com/source").
			End(),
		"data, attributes and extensions and specversion with struct data v0.3": new(orderedJsonObjectBuilder).Start().
			Add("data", map[string]interface{}{
				"a": 42,
				"b": "testing",
			}).
			Add("id", "ABC-123").
			Add("time", now.Format(time.RFC3339Nano)).
			Add("type", "com.example.test").
			Add("exbool", true).
			Add("exint", 42).
			Add("exstring", "exstring").
			Add("exbinary", "AAECAw==").
			Add("exurl", "http://example.com/source").
			Add("extime", now.Format(time.RFC3339Nano)).
			Add("schemaurl", "http://example.com/schema").
			Add("source", "http://example.com/source").
			Add("datacontenttype", "application/json").
			Add("specversion", "0.3").
			End(),
		"struct data v1.0": new(orderedJsonObjectBuilder).Start().
			Add("specversion", "1.0").
			Add("datacontenttype", "application/json").
			Add("data", map[string]interface{}{
				"a": 42,
				"b": "testing",
			}).
			Add("id", "ABC-123").
			Add("time", now.Format(time.RFC3339Nano)).
			Add("type", "com.example.test").
			Add("exbool", true).
			Add("exint", 42).
			Add("exstring", "exstring").
			Add("exbinary", "AAECAw==").
			Add("exurl", "http://example.com/source").
			Add("extime", now.Format(time.RFC3339Nano)).
			Add("dataschema", "http://example.com/schema").
			Add("source", "http://example.com/source").
			End(),
		"data, attributes and extensions and specversion with struct data v1.0": new(orderedJsonObjectBuilder).Start().
			Add("data", map[string]interface{}{
				"a": 42,
				"b": "testing",
			}).
			Add("id", "ABC-123").
			Add("time", now.Format(time.RFC3339Nano)).
			Add("type", "com.example.test").
			Add("exbool", true).
			Add("exint", 42).
			Add("exstring", "exstring").
			Add("exbinary", "AAECAw==").
			Add("exurl", "http://example.com/source").
			Add("extime", now.Format(time.RFC3339Nano)).
			Add("dataschema", "http://example.com/schema").
			Add("source", "http://example.com/source").
			Add("datacontenttype", "application/json").
			Add("specversion", "1.0").
			End(),
		"string data v1.0": new(orderedJsonObjectBuilder).Start().
			Add("specversion", "1.0").
			Add("datacontenttype", "application/json").
			Add("data", "This is a string.").
			Add("id", "ABC-123").
			Add("time", now.Format(time.RFC3339Nano)).
			Add("type", "com.example.test").
			Add("exbool", true).
			Add("exint", 42).
			Add("exstring", "exstring").
			Add("exbinary", "AAECAw==").
			Add("exurl", "http://example.com/source").
			Add("extime", now.Format(time.RFC3339Nano)).
			Add("dataschema", "http://example.com/schema").
			Add("source", "http://example.com/source").
			End(),
		"base64 json encoded data v1.0": new(orderedJsonObjectBuilder).Start().
			Add("specversion", "1.0").
			Add("datacontenttype", "application/json").
			Add("data_base64", base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`))).
			Add("id", "ABC-123").
			Add("time", now.Format(time.RFC3339Nano)).
			Add("type", "com.example.test").
			Add("dataschema", "http://example.com/schema").
			Add("source", "http://example.com/source").
			End(),
		"base64 xml encoded data v1.0": new(orderedJsonObjectBuilder).Start().
			Add("specversion", "1.0").
			Add("datacontenttype", "application/json").
			Add("data_base64", base64.StdEncoding.EncodeToString(mustEncodeWithDataCodec(b, event.ApplicationXML, &XMLDataExample{AnInt: 10}))).
			Add("id", "ABC-123").
			Add("time", now.Format(time.RFC3339Nano)).
			Add("type", "com.example.test").
			Add("dataschema", "http://example.com/schema").
			Add("source", "http://example.com/source").
			End(),
		"nil data v1.0": new(orderedJsonObjectBuilder).Start().
			Add("specversion", "1.0").
			Add("datacontenttype", "application/json").
			Add("id", "ABC-123").
			Add("time", now.Format(time.RFC3339Nano)).
			Add("type", "com.example.test").
			Add("exbool", true).
			Add("exint", 42).
			Add("exstring", "exstring").
			Add("exbinary", "AAECAw==").
			Add("exurl", "http://example.com/source").
			Add("extime", now.Format(time.RFC3339Nano)).
			Add("dataschema", "http://example.com/schema").
			Add("source", "http://example.com/source").
			End(),
	}
	for n, tc := range testCases {
		bytes := []byte(tc)
		b.Run(n, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Event = event.Event{}
				Error = json.Unmarshal(bytes, &Event)
			}
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
