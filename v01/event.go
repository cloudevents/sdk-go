package v01

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/dispatchframework/cloudevents-go-sdk"
)

const (
	eventTypeKey        = "eventType"
	eventTypeVersionKey = "eventTypeVersion"
	sourceKey           = "sourceKey"
	eventIDKey          = "eventID"
	eventTimeKey        = "eventTime"
	schemaURLKey        = "schemaURL"
	contentTypeKey      = "contentType"
	extensionsKey       = "extensions"
	dataKey             = "data"
)

// Event implements the the CloudEvents specification version 0.1
// https://github.com/cloudevents/spec/blob/v0.1/spec.md
type Event struct {
	// EventType is a mandatory property
	// https://github.com/cloudevents/spec/blob/v0.1/spec.md#eventtype
	EventType string `json:"eventType"`
	// EventTypeVersion is an optional property
	// https://github.com/cloudevents/spec/blob/v0.1/spec.md#eventtypeversion
	EventTypeVersion string `json:"eventTypeVersion,omitempty"`
	// Source is a mandatory property
	// TODO: ensure URI parsing
	// https://github.com/cloudevents/spec/blob/v0.1/spec.md#source
	Source string `json:"source"`
	// EventID is a mandatory property
	// https://github.com/cloudevents/spec/blob/v0.1/spec.md#eventid
	EventID string `json:"eventID"`
	// EventTime is an optional property
	// https://github.com/cloudevents/spec/blob/v0.1/spec.md#eventtime
	EventTime *time.Time `json:"eventTime,omitempty"`
	// SchemaURL is an optional property
	// https://github.com/cloudevents/spec/blob/v0.1/spec.md#schemaurl
	SchemaURL string `json:"schemaURL,omitempty"`
	// ContentType is an optional property
	// https://github.com/cloudevents/spec/blob/v0.1/spec.md#contenttype
	ContentType string `json:"contentType,omitempty"`
	// Extensions is an optional property
	// https://github.com/cloudevents/spec/blob/v0.1/spec.md#extensions
	Extensions map[string]interface{} `json:"extensions,omitempty"`
	// Data is an optional property
	// https://github.com/cloudevents/spec/blob/v0.1/spec.md#data-1
	Data interface{} `json:"data,omitempty"`
}

// CloudEventVersion returns the CloudEvents specification version supported by this implementation
func (Event) CloudEventVersion() (version string) {
	return cloudevents.Version01
}

// Properties returns the map of all supported properties in version 0.1.
// The map value says whether particular property is required.
func (Event) Properties() map[string]bool {
	return map[string]bool{
		eventTypeKey:        true,
		sourceKey:           true,
		eventIDKey:          true,
		eventTypeVersionKey: false,
		eventTimeKey:        false,
		schemaURLKey:        false,
		contentTypeKey:      false,
		extensionsKey:       false,
		dataKey:             false,
	}
}

// Get implements a generic getter method
func (e *Event) Get(property string) (interface{}, bool) {
	field := reflect.ValueOf(e).Elem().FieldByName(strings.Title(property))
	if field.IsValid() {
		return field.Interface(), true
	}
	if e.Extensions == nil {
		return nil, false
	}
	if value, ok := e.Extensions[property]; ok {
		return value, ok
	}
	return nil, false
}

// Set sets the arbitrary property of event.
func (e *Event) Set(property string, value interface{}) {
	field := reflect.ValueOf(e).Elem().FieldByName(strings.Title(property))
	if field.IsValid() {
		field.Set(reflect.ValueOf(value))
		return
	}

	if e.Extensions == nil {
		e.Extensions = make(map[string]interface{})
	}
	e.Extensions[property] = value
}

// Validate returns an error if the event is not correct according to the spec.
func (e *Event) Validate() error {
	if e.EventType == "" {
		return cloudevents.RequiredPropertyError(eventTypeKey)
	}
	if e.EventID == "" {
		return cloudevents.RequiredPropertyError(eventIDKey)
	}
	if e.Source == "" {
		return cloudevents.RequiredPropertyError(sourceKey)
	}
}

// MarshalJSON implements the JSON Marshaler interface.
func (e *Event) MarshalJSON() ([]byte, error) {
	type tmp Event
	eventWithVersion := struct {
		CloudEventVersion string `json:"cloudEventVersion"`
		*tmp
	}{
		CloudEventVersion: cloudevents.Version01,
		tmp:               (*tmp)(e),
	}
	return json.Marshal(&eventWithVersion)
}
