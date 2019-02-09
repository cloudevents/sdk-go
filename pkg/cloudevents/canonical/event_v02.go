package canonical

import (
	"strings"
)

const (
	// CloudEventsVersionV02 represents the version 0.2 of the CloudEvents spec.
	CloudEventsVersionV02 = "0.2"
)

// EventContextV02 represents the non-data attributes of a CloudEvents v0.2
// event.
type EventContextV02 struct {
	// The version of the CloudEvents specification used by the event.
	SpecVersion string `json:"specversion"`
	// The type of the occurrence which has happened.
	Type string `json:"type"`
	// A URI describing the event producer.
	Source URLRef `json:"source"`
	// ID of the event; must be non-empty and unique within the scope of the producer.
	ID string `json:"id"`
	// Timestamp when the event happened.
	Time *Timestamp `json:"time,omitempty"`
	// A link to the schema that the `data` attribute adheres to.
	SchemaURL *URLRef `json:"schemaurl,omitempty"`
	// A MIME (RFC2046) string describing the media type of `data`.
	// TODO: Should an empty string assume `application/json`, `application/octet-stream`, or auto-detect the content?
	ContentType string `json:"contenttype,omitempty"`
	// Additional extension metadata beyond the base spec.
	Extensions map[string]interface{} `json:"-,omitempty"`
}

var _ EventContext = (*EventContextV02)(nil)

// DataContentType implements the StructuredSender interface.
func (ec EventContextV02) DataContentType() string {
	return ec.ContentType
}

// AsV01 implements the ContextTranslator interface.
func (ec EventContextV02) AsV01() EventContextV01 {
	ret := EventContextV01{
		CloudEventsVersion: CloudEventsVersionV01,
		EventID:            ec.ID,
		EventTime:          ec.Time,
		EventType:          ec.Type,
		SchemaURL:          ec.SchemaURL,
		ContentType:        ec.ContentType,
		Source:             ec.Source,
		Extensions:         make(map[string]interface{}),
	}
	for k, v := range ec.Extensions {
		// eventTypeVersion was retired in v0.2
		if strings.EqualFold(k, "eventTypeVersion") {
			etv, ok := v.(string)
			if ok {
				ret.EventTypeVersion = etv
			}
			continue
		}
		ret.Extensions[k] = v
	}
	if len(ret.Extensions) == 0 {
		ret.Extensions = nil
	}
	return ret
}

// AsV02 implements the ContextTranslator interface.
func (ec EventContextV02) AsV02() EventContextV02 {
	return ec
}
