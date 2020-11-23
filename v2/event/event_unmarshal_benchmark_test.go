package event_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

var Event event.Event

func BenchmarkUnmarshal(b *testing.B) {
	now := types.Timestamp{Time: time.Now().UTC()}

	testCases := map[string][]byte{
		"struct data v0.3": mustJsonMarshal(b, map[string]interface{}{
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
		"string data v0.3": mustJsonMarshal(b, map[string]interface{}{
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
		"nil data v0.3": mustJsonMarshal(b, map[string]interface{}{
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
		"data, attributes and extensions and specversion with struct data v0.3": mustJsonMarshal(b, map[string]interface{}{
			"data": map[string]interface{}{
				"a": 42,
				"b": "testing",
			},
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
			"datacontenttype": "application/json",
			"specversion":     "0.3",
		}),
		"struct data v1.0": mustJsonMarshal(b, map[string]interface{}{
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
		"data, attributes and extensions and specversion with struct data v1.0": mustJsonMarshal(b, map[string]interface{}{
			"data": map[string]interface{}{
				"a": 42,
				"b": "testing",
			},
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
			"datacontenttype": "application/json",
			"specversion":     "1.0",
		}),
		"string data v1.0": mustJsonMarshal(b, map[string]interface{}{
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
		"base64 json encoded data v1.0": mustJsonMarshal(b, map[string]interface{}{
			"specversion":     "1.0",
			"datacontenttype": "application/json",
			"data_base64":     base64.StdEncoding.EncodeToString([]byte(`{"hello":"world"}`)),
			"id":              "ABC-123",
			"time":            now.Format(time.RFC3339Nano),
			"type":            "com.example.test",
			"dataschema":      "http://example.com/schema",
			"source":          "http://example.com/source",
		}),
		"base64 xml encoded data v1.0": mustJsonMarshal(b, map[string]interface{}{
			"specversion":     "1.0",
			"datacontenttype": "application/json",
			"data_base64":     base64.StdEncoding.EncodeToString(mustEncodeWithDataCodec(b, event.ApplicationXML, &XMLDataExample{AnInt: 10})),
			"id":              "ABC-123",
			"time":            now.Format(time.RFC3339Nano),
			"type":            "com.example.test",
			"dataschema":      "http://example.com/schema",
			"source":          "http://example.com/source",
		}),
		"nil data v1.0": mustJsonMarshal(b, map[string]interface{}{
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
	}
	for n, tc := range testCases {
		bytes := tc
		b.Run(n, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Event = event.Event{}
				Error = json.Unmarshal(bytes, &Event)
			}
		})
	}
}

// This is a little hack we need to create a json ordered.
// This makes the bench reproducible
type orderedJsonObjectBuilder strings.Builder

func (b *orderedJsonObjectBuilder) Start() {
	(*strings.Builder)(b).WriteRune('{')
}

func (b *orderedJsonObjectBuilder) Add(key string, value interface{}) {
	(*strings.Builder)(b).WriteRune('"')
	(*strings.Builder)(b).WriteString(key)
	(*strings.Builder)(b).WriteString("\":")

	v, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	(*strings.Builder)(b).Write(v)
	(*strings.Builder)(b).WriteRune(',')
}

func (b *orderedJsonObjectBuilder) End() {
	(*strings.Builder)(b).WriteRune('}')
}
