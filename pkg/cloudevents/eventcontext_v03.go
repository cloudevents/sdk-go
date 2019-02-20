package cloudevents

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

// WIP: AS OF FEB 19, 2019

const (
	// CloudEventsVersionV03 represents the version 0.3 of the CloudEvents spec.
	CloudEventsVersionV03 = "0.3"
)

// EventContextV03 represents the non-data attributes of a CloudEvents v0.3
// event.
type EventContextV03 struct {
	// SpecVersion - The version of the CloudEvents specification used by the event.
	SpecVersion string `json:"specversion"`
	// Type - The type of the occurrence which has happened.
	Type string `json:"type"`
	// Source - A URI describing the event producer.
	Source types.URLRef `json:"source"`
	// ID of the event; must be non-empty and unique within the scope of the producer.
	ID string `json:"id"`
	// Time - A Timestamp when the event happened.
	Time *types.Timestamp `json:"time,omitempty"`
	// SchemaURL - A link to the schema that the `data` attribute adheres to.
	SchemaURL *types.URLRef `json:"schemaurl,omitempty"`
	// GetDataContentType - A MIME (RFC2046) string describing the media type of `data`.
	// TODO: Should an empty string assume `application/json`, `application/octet-stream`, or auto-detect the content?
	DataContentType string `json:"datacontenttype,omitempty"`
	// Extensions - Additional extension metadata beyond the base spec.
	Extensions map[string]interface{} `json:"-,omitempty"` // TODO: decide how we want extensions to be inserted
}

var _ EventContext = (*EventContextV03)(nil)

func (ec EventContextV03) GetSpecVersion() string {
	if ec.SpecVersion != "" {
		return ec.SpecVersion
	}
	return CloudEventsVersionV03
}

func (ec EventContextV03) GetDataContentType() string {
	return ec.DataContentType
}

func (ec EventContextV03) GetType() string {
	return ec.Type
}

func (ec EventContextV03) AsV01() EventContextV01 {
	ecv2 := ec.AsV02()
	return ecv2.AsV01()
}

func (ec EventContextV03) AsV02() EventContextV02 {
	ret := EventContextV02{
		SpecVersion: CloudEventsVersionV02,
		ID:          ec.ID,
		Time:        ec.Time,
		Type:        ec.Type,
		SchemaURL:   ec.SchemaURL,
		ContentType: ec.DataContentType,
		Source:      ec.Source,
		Extensions:  ec.Extensions,
	}
	return ret
}

func (ec EventContextV03) AsV03() EventContextV03 {
	ec.SpecVersion = CloudEventsVersionV03
	return ec
}
