package extensions

import (
	"reflect"
	"strings"

	"github.com/lightstep/tracecontext.go/traceparent"
	"go.opencensus.io/trace"
)

// EventTracer interface allows setting extension for cloudevents context.
type EventTracer interface {
	SetExtension(k string, v interface{}) error
}

// DistributedTracingExtension represents the extension for cloudevents context
type DistributedTracingExtension struct {
	TraceParent string `json:"traceparent"`
	TraceState  string `json:"tracestate"`
}

// AddTracingAttributes adds the tracing attributes traceparent and tracestate to the cloudevents context
func (d DistributedTracingExtension) AddTracingAttributes(ec EventTracer) error {
	if d.TraceParent != "" {
		value := reflect.ValueOf(d)
		typeOf := value.Type()

		for i := 0; i < value.NumField(); i++ {
			k := strings.ToLower(typeOf.Field(i).Name)
			v := value.Field(i).Interface()
			if k == "tracestate" && v == "" {
				continue
			}
			if err := ec.SetExtension(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// FromSpanContext populates DistributedTracingExtension from a SpanContext.
func FromSpanContext(sc trace.SpanContext) DistributedTracingExtension {
	tp := traceparent.TraceParent{
		TraceID: sc.TraceID,
		SpanID:  sc.SpanID,
		Flags: traceparent.Flags{
			Recorded: sc.IsSampled(),
		},
	}

	var entries = make([]string, 0, len(sc.Tracestate.Entries()))
	for _, entry := range sc.Tracestate.Entries() {
		entries = append(entries, strings.Join([]string{entry.Key, entry.Value}, "="))
	}

	return DistributedTracingExtension{
		TraceParent: tp.String(),
		TraceState:  strings.Join(entries, ","),
	}
}
