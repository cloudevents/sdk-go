package cloudevents

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"strings"
)

const (
	// CloudEventsVersionV01 represents the version 0.1 of the CloudEvents spec.
	CloudEventsVersionV01 = "0.1"
)

// EventContextV01 holds standard metadata about an event. See
// https://github.com/cloudevents/spec/blob/v0.1/spec.md#context-attributes for
// details on these fields.
type EventContextV01 struct {
	// The version of the CloudEvents specification used by the event.
	CloudEventsVersion string `json:"cloudEventsVersion,omitempty"`
	// ID of the event; must be non-empty and unique within the scope of the producer.
	EventID string `json:"eventID"`
	// Timestamp when the event happened.
	EventTime *types.Timestamp `json:"eventTime,omitempty"`
	// Type of occurrence which has happened.
	EventType string `json:"eventType"`
	// The version of the `eventType`; this is producer-specific.
	EventTypeVersion string `json:"eventTypeVersion,omitempty"`
	// A link to the schema that the `data` attribute adheres to.
	SchemaURL *types.URLRef `json:"schemaURL,omitempty"`
	// A MIME (RFC 2046) string describing the media type of `data`.
	// TODO: Should an empty string assume `application/json`, or auto-detect the content?
	ContentType string `json:"contentType,omitempty"`
	// A URI describing the event producer.
	Source types.URLRef `json:"source"`
	// Additional metadata without a well-defined structure.
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

var _ EventContext = (*EventContextV01)(nil)

func (ec EventContextV01) GetSpecVersion() string {
	if ec.CloudEventsVersion != "" {
		return ec.CloudEventsVersion
	}
	return CloudEventsVersionV01
}

func (ec EventContextV01) GetDataContentType() string {
	// TODO: there are cases where there is char encoding info on the content type.
	// Fix this for these cases as we find them.
	if strings.HasSuffix(ec.ContentType, "json") {
		return "application/json"
	}
	if strings.HasSuffix(ec.ContentType, "xml") {
		return "application/xml"
	}
	return ec.ContentType
}

func (ec EventContextV01) GetType() string {
	return ec.EventType
}

func (ec EventContextV01) AsV01() EventContextV01 {
	ec.CloudEventsVersion = CloudEventsVersionV01
	return ec
}

func (ec EventContextV01) AsV02() EventContextV02 {
	ret := EventContextV02{
		SpecVersion: CloudEventsVersionV02,
		Type:        ec.EventType,
		Source:      ec.Source,
		ID:          ec.EventID,
		Time:        ec.EventTime,
		SchemaURL:   ec.SchemaURL,
		ContentType: ec.ContentType,
		Extensions:  make(map[string]interface{}),
	}
	// eventTypeVersion was retired in v0.2, so put it in an extension.
	if ec.EventTypeVersion != "" {
		ret.Extensions["eventTypeVersion"] = ec.EventTypeVersion
	}
	if ec.Extensions != nil {
		for k, v := range ec.Extensions {
			ret.Extensions[k] = v
		}
	}
	if len(ret.Extensions) == 0 {
		ret.Extensions = nil
	}
	return ret
}

func (ec EventContextV01) AsV03() EventContextV03 {
	ecv2 := ec.AsV02()
	return ecv2.AsV03()
}
