package extensions_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/extensions"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
)

type Data struct {
	Message string
}

var now = types.Timestamp{Time: time.Now().UTC()}

var sourceUrl, _ = url.Parse("http://example.com/source")
var source = &types.URLRef{URL: *sourceUrl}

var schemaUrl, _ = url.Parse("http://example.com/schema")
var schema = &types.URLRef{URL: *schemaUrl}

type values struct {
	context interface{}
	want    map[string]interface{}
}

func TestAddTracingAttributes_Scenario1(t *testing.T) {
	var st = extensions.DistributedTracingExtension{
		TraceParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		TraceState:  "rojo=00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01,congo=lZWRzIHRoNhcm5hbCBwbGVhc3VyZS4=",
	}

	var eventContextVersions = map[string]values{
		"EventContextV01": {
			context: cloudevents.EventContextV01{
				EventID:     "ABC-123",
				EventTime:   &now,
				EventType:   "com.example.test",
				SchemaURL:   schema,
				ContentType: cloudevents.StringOfApplicationJSON(),
				Source:      *source,
			},
			want: map[string]interface{}{"traceparent": st.TraceParent, "tracestate": st.TraceState},
		},
		"EventContextV02": {
			context: cloudevents.EventContextV02{
				ID:          "ABC-123",
				Time:        &now,
				Type:        "com.example.test",
				SchemaURL:   schema,
				ContentType: cloudevents.StringOfApplicationJSON(),
				Source:      *source,
			},
			want: map[string]interface{}{"traceparent": st.TraceParent, "tracestate": st.TraceState},
		},
		"EventContextV03": {
			context: cloudevents.EventContextV03{
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.test",
				SchemaURL:       schema,
				DataContentType: cloudevents.StringOfApplicationJSON(),
				Source:          *source,
			},
			want: map[string]interface{}{"traceparent": st.TraceParent, "tracestate": st.TraceState},
		},
	}

	for k, ecv := range eventContextVersions {
		testAddTracingAttributesFunc(t, st, ecv, k)
	}
}

func TestAddTracingAttributes_Scenario2(t *testing.T) {
	var st = extensions.DistributedTracingExtension{
		TraceParent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
	}

	var eventContextVersions = map[string]values{
		"EventContextV01": {
			context: cloudevents.EventContextV01{
				EventID:     "ABC-123",
				EventTime:   &now,
				EventType:   "com.example.test",
				SchemaURL:   schema,
				ContentType: cloudevents.StringOfApplicationJSON(),
				Source:      *source,
			},
			want: map[string]interface{}{"traceparent": st.TraceParent},
		},
		"EventContextV02": {
			context: cloudevents.EventContextV02{
				ID:          "ABC-123",
				Time:        &now,
				Type:        "com.example.test",
				SchemaURL:   schema,
				ContentType: cloudevents.StringOfApplicationJSON(),
				Source:      *source,
			},
			want: map[string]interface{}{"traceparent": st.TraceParent},
		},
		"EventContextV03": {
			context: cloudevents.EventContextV03{
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.test",
				SchemaURL:       schema,
				DataContentType: cloudevents.StringOfApplicationJSON(),
				Source:          *source,
			},
			want: map[string]interface{}{"traceparent": st.TraceParent},
		},
	}

	for k, ecv := range eventContextVersions {
		testAddTracingAttributesFunc(t, st, ecv, k)
	}
}

func TestAddTracingAttributes_Scenario3(t *testing.T) {
	var st = extensions.DistributedTracingExtension{}

	var eventContextVersions = map[string]values{
		"EventContextV01": {
			context: cloudevents.EventContextV01{
				EventID:     "ABC-123",
				EventTime:   &now,
				EventType:   "com.example.test",
				SchemaURL:   schema,
				ContentType: cloudevents.StringOfApplicationJSON(),
				Source:      *source,
			},
			want: map[string]interface{}(nil),
		},
		"EventContextV02": {
			context: cloudevents.EventContextV02{
				ID:          "ABC-123",
				Time:        &now,
				Type:        "com.example.test",
				SchemaURL:   schema,
				ContentType: cloudevents.StringOfApplicationJSON(),
				Source:      *source,
			},
			want: map[string]interface{}(nil),
		},
		"EventContextV03": {
			context: cloudevents.EventContextV03{
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.test",
				SchemaURL:       schema,
				DataContentType: cloudevents.StringOfApplicationJSON(),
				Source:          *source,
			},
			want: map[string]interface{}(nil),
		},
	}

	for k, ecv := range eventContextVersions {
		testAddTracingAttributesFunc(t, st, ecv, k)
	}
}

func TestAddTracingAttributes_Scenario4(t *testing.T) {
	var st = extensions.DistributedTracingExtension{
		TraceState: "rojo=00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01,congo=lZWRzIHRoNhcm5hbCBwbGVhc3VyZS4=",
	}

	var eventContextVersions = map[string]values{
		"EventContextV01": {
			context: cloudevents.EventContextV01{
				EventID:     "ABC-123",
				EventTime:   &now,
				EventType:   "com.example.test",
				SchemaURL:   schema,
				ContentType: cloudevents.StringOfApplicationJSON(),
				Source:      *source,
			},
			want: map[string]interface{}(nil),
		},
		"EventContextV02": {
			context: cloudevents.EventContextV02{
				ID:          "ABC-123",
				Time:        &now,
				Type:        "com.example.test",
				SchemaURL:   schema,
				ContentType: cloudevents.StringOfApplicationJSON(),
				Source:      *source,
			},
			want: map[string]interface{}(nil),
		},
		"EventContextV03": {
			context: cloudevents.EventContextV03{
				ID:              "ABC-123",
				Time:            &now,
				Type:            "com.example.test",
				SchemaURL:       schema,
				DataContentType: cloudevents.StringOfApplicationJSON(),
				Source:          *source,
			},
			want: map[string]interface{}(nil),
		},
	}

	for k, ecv := range eventContextVersions {
		testAddTracingAttributesFunc(t, st, ecv, k)
	}
}

func testAddTracingAttributesFunc(t *testing.T, st extensions.DistributedTracingExtension, ecv values, ces string) {
	var event cloudevents.Event
	switch ces {
	case "EventContextV01":
		ectx := ecv.context.(cloudevents.EventContextV01).AsV01()
		st.AddTracingAttributes(ectx)
		event = cloudevents.Event{Context: ectx, Data: &Data{Message: "Hello world"}}
	case "EventContextV02":
		ectx := ecv.context.(cloudevents.EventContextV02).AsV02()
		st.AddTracingAttributes(ectx)
		event = cloudevents.Event{Context: ectx, Data: &Data{Message: "Hello world"}}
	case "EventContextV03":
		ectx := ecv.context.(cloudevents.EventContextV03).AsV03()
		st.AddTracingAttributes(ectx)
		event = cloudevents.Event{Context: ectx, Data: &Data{Message: "Hello world"}}
	}
	got := event.Extensions()

	if diff := cmp.Diff(ecv.want, got); diff != "" {
		t.Errorf("\nunexpected (-want, +got) = %v", diff)
	}
}
