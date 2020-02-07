package extensions

import (
	"reflect"
	"strings"
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
