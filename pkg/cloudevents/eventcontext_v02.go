package cloudevents

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
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
	Source types.URLRef `json:"source"`
	// ID of the event; must be non-empty and unique within the scope of the producer.
	ID string `json:"id"`
	// Timestamp when the event happened.
	Time *types.Timestamp `json:"time,omitempty"`
	// A link to the schema that the `data` attribute adheres to.
	SchemaURL *types.URLRef `json:"schemaurl,omitempty"`
	// A MIME (RFC2046) string describing the media type of `data`.
	// TODO: Should an empty string assume `application/json`, `application/octet-stream`, or auto-detect the content?
	ContentType string `json:"contenttype,omitempty"`
	// Additional extension metadata beyond the base spec.
	Extensions map[string]interface{} `json:"-,omitempty"` // TODO: decide how we want extensions to be inserted
}

var _ EventContext = (*EventContextV02)(nil)

func (ec EventContextV02) GetSpecVersion() string {
	if ec.SpecVersion != "" {
		return ec.SpecVersion
	}
	return CloudEventsVersionV02
}

func (ec EventContextV02) GetDataContentType() string {
	return ec.ContentType
}

func (ec EventContextV02) GetType() string {
	return ec.Type
}

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

func (ec EventContextV02) AsV02() EventContextV02 {
	ec.SpecVersion = CloudEventsVersionV02
	return ec
}

func (ec EventContextV02) AsV03() EventContextV03 {
	ret := EventContextV03{
		SpecVersion:     CloudEventsVersionV03,
		ID:              ec.ID,
		Time:            ec.Time,
		Type:            ec.Type,
		SchemaURL:       ec.SchemaURL,
		DataContentType: ec.ContentType,
		Source:          ec.Source,
		Extensions:      ec.Extensions,
	}
	return ret
}
