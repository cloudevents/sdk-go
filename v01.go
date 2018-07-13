package cloudevents

import (
	"fmt"
	"time"
)

// Version01 defines a version string for CloudEvents specification version 0.1. See also EventV01
// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/spec.md
const Version01 = "0.1"

// GetEventV01 converts CloudEvent to version 0.1, if possible. If not, returns an error.
func GetEventV01(event CloudEvent) (*EventV01, error) {
	if val, ok := event.(*EventV01); ok {
		return val, nil
	}
	errorMsg := fmt.Sprintf("CloudEvent reports version %s, but is not using EventV01 structure", event.CloudEventVersion())
	return nil, VersionMismatchError(errorMsg)
}

// EventV01 implements Cloud Events specification version 0.1
// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/spec.md
type EventV01 struct {
	EventType        string                 `json:"eventType"`
	EventTypeVersion string                 `json:"eventTypeVersion"`
	Source           string                 `json:"source"`
	EventID          string                 `json:"eventID"`
	EventTime        time.Time              `json:"eventTime"`
	SchemaURL        string                 `json:"schemaURL"`
	ContentType      string                 `json:"contentType"`
	Extensions       map[string]interface{} `json:"extensions"`
	Data             interface{}            `json:"data"`
}

// CloudEventVersion returns the version string this implementation follows.
func (EventV01) CloudEventVersion() string {
	return Version01
}

// Context returns a map with metadata properties defined by the event.
func (e *EventV01) Context() map[string]interface{} {
	return map[string]interface{}{
		"eventType":          e.EventType,
		"eventTypeVersion":   e.EventTypeVersion,
		"cloudEventsVersion": Version01,
		"source":             e.Source,
		"eventID":            e.EventID,
		"eventTime":          e.EventTime,
		"schemaURL":          e.SchemaURL,
		"contentType":        e.ContentType,
		"extensions":         e.Extensions,
	}
}

// UnmarshalJSON implements JSON Unmarshaler for CloudEvents version 0.1, following the spec as described in:
// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/json-format.md
func (e *EventV01) UnmarshalJSON([]byte) error {
	panic("implement me")
}

// MarshalJSON implements JSON Marshaler for CloudEvents version 0.1, following the spec as described in:
// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/json-format.md
func (e *EventV01) MarshalJSON() ([]byte, error) {
	panic("implement me")
}
