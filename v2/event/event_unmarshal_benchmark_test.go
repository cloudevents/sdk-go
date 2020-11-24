package event_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

var Event event.Event

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
