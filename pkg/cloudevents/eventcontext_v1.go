package cloudevents

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

// WIP: AS OF SEP 20, 2019

const (
	// CloudEventsVersionV1 represents the version 1.0 of the CloudEvents spec.
	CloudEventsVersionV1 = "1.0"
)

// EventContextV1 represents the non-data attributes of a CloudEvents v0.3
// event.
type EventContextV1 struct {
	// SpecVersion - The version of the CloudEvents specification used by the event.
	SpecVersion string `json:"specversion"`
	// Type - The type of the occurrence which has happened.
	Type string `json:"type"`
	// Source - A URI describing the event producer.
	Source types.URLRef `json:"source"`
	// Subject - The subject of the event in the context of the event producer
	// (identified by `source`).
	Subject *string `json:"subject,omitempty"`
	// ID of the event; must be non-empty and unique within the scope of the producer.
	ID string `json:"id"`
	// Time - A Timestamp when the event happened.
	Time *types.Timestamp `json:"time,omitempty"`
	// DataSchema - A link to the schema that the `data` attribute adheres to.
	DataSchema *types.URLRef `json:"dataschema,omitempty"` // TODO: spec changed to URL.
	// GetDataMediaType - A MIME (RFC2046) string describing the media type of `data`.
	// TODO: Should an empty string assume `application/json`, `application/octet-stream`, or auto-detect the content?
	DataContentType *string `json:"datacontenttype,omitempty"`
	// Extensions - Additional extension metadata beyond the base spec.
	Extensions map[string]string `json:"-"`
}

// Adhere to EventContext
var _ EventContext = (*EventContextV1)(nil)

// ExtensionAs implements EventContext.ExtensionAs
func (ec EventContextV1) ExtensionAs(name string, obj interface{}) error {
	value, ok := ec.Extensions[name]
	if !ok {
		return fmt.Errorf("extension %q does not exist", name)
	}

	// Only support *string for now.
	switch v := obj.(type) {
	case *string:
		*v = value
		return nil
	default:
		return fmt.Errorf("unknown extension type %T", obj)
	}
}

// SetExtension adds the extension 'name' with value 'value' to the CloudEvents context.
func (ec *EventContextV1) SetExtension(name string, value interface{}) error {
	if ec.Extensions == nil {
		ec.Extensions = make(map[string]string)
	}
	if value == nil {
		delete(ec.Extensions, name)
	} else {
		ec.Extensions[name] = fmt.Sprintf("%s", value) // TODO we might need to do something about encoding the string.
	}
	return nil
}

// Clone implements EventContextConverter.Clone
func (ec EventContextV1) Clone() EventContext {
	return ec.AsV1()
}

// AsV01 implements EventContextConverter.AsV01
func (ec EventContextV1) AsV01() *EventContextV01 {
	ecv2 := ec.AsV02()
	return ecv2.AsV01()
}

// AsV02 implements EventContextConverter.AsV02
func (ec EventContextV1) AsV02() *EventContextV02 {
	ecv3 := ec.AsV03()
	return ecv3.AsV02()
}

// AsV03 implements EventContextConverter.AsV03
func (ec EventContextV1) AsV03() *EventContextV03 {
	ret := EventContextV03{
		SpecVersion:     CloudEventsVersionV02,
		ID:              ec.ID,
		Time:            ec.Time,
		Type:            ec.Type,
		SchemaURL:       ec.DataSchema,
		DataContentType: ec.DataContentType,
		//DeprecatedDataContentEncoding: ec.DeprecatedDataContentEncoding, // TODO fix up DeprecatedDataContentEncoding
		Source:     ec.Source,
		Subject:    ec.Subject,
		Extensions: make(map[string]interface{}),
	}
	// TODO: DeprecatedDataContentEncoding needs to be moved to extensions.
	if ec.Extensions != nil {
		for k, v := range ec.Extensions {
			ret.Extensions[k] = v
		}
	}
	if len(ret.Extensions) == 0 {
		ret.Extensions = nil
	}
	return &ret
}

// AsV04 implements EventContextConverter.AsV04
func (ec EventContextV1) AsV1() *EventContextV1 {
	ec.SpecVersion = CloudEventsVersionV1
	return &ec
}

// Validate returns errors based on requirements from the CloudEvents spec.
// For more details, see https://github.com/cloudevents/spec/blob/master/spec.md
// As of Feb 26, 2019, commit
//
// TODO: UPDATE THIS FOR v1.0
//
// + https://github.com/cloudevents/spec/pull/TODO -> extensions change
// + https://github.com/cloudevents/spec/pull/TODO -> dataschema
func (ec EventContextV1) Validate() error {
	errors := []string(nil)

	// TODO: a lot of these have changed. Double check them all.

	// type
	// Type: String
	// Constraints:
	//  REQUIRED
	//  MUST be a non-empty string
	//  SHOULD be prefixed with a reverse-DNS name. The prefixed domain dictates the organization which defines the semantics of this event type.
	eventType := strings.TrimSpace(ec.Type)
	if eventType == "" {
		errors = append(errors, "type: MUST be a non-empty string")
	}

	// specversion
	// Type: String
	// Constraints:
	//  REQUIRED
	//  MUST be a non-empty string
	specVersion := strings.TrimSpace(ec.SpecVersion)
	if specVersion == "" {
		errors = append(errors, "specversion: MUST be a non-empty string")
	}

	// source
	// Type: URI-reference
	// Constraints:
	//  REQUIRED
	source := strings.TrimSpace(ec.Source.String())
	if source == "" {
		errors = append(errors, "source: REQUIRED")
	}

	// subject
	// Type: String
	// Constraints:
	//  OPTIONAL
	//  MUST be a non-empty string
	if ec.Subject != nil {
		subject := strings.TrimSpace(*ec.Subject)
		if subject == "" {
			errors = append(errors, "subject: if present, MUST be a non-empty string")
		}
	}

	// id
	// Type: String
	// Constraints:
	//  REQUIRED
	//  MUST be a non-empty string
	//  MUST be unique within the scope of the producer
	id := strings.TrimSpace(ec.ID)
	if id == "" {
		errors = append(errors, "id: MUST be a non-empty string")

		// no way to test "MUST be unique within the scope of the producer"
	}

	// time
	// Type: Timestamp
	// Constraints:
	//  OPTIONAL
	//  If present, MUST adhere to the format specified in RFC 3339
	// --> no need to test this, no way to set the time without it being valid.

	// dataschema
	// Type: URI
	// Constraints:
	//  OPTIONAL
	//  If present, MUST adhere to the format specified in RFC 3986
	if ec.DataSchema != nil {
		dataSchema := strings.TrimSpace(ec.DataSchema.String())
		// empty string is not RFC 3986 compatible.
		if dataSchema == "" {
			errors = append(errors, "dataschema: if present, MUST adhere to the format specified in RFC 3986")
		}
	}

	// datacontenttype
	// Type: String per RFC 2046
	// Constraints:
	//  OPTIONAL
	//  If present, MUST adhere to the format specified in RFC 2046
	if ec.DataContentType != nil {
		dataContentType := strings.TrimSpace(*ec.DataContentType)
		if dataContentType == "" {
			// TODO: need to test for RFC 2046
			errors = append(errors, "datacontenttype: if present, MUST adhere to the format specified in RFC 2046")
		}
	}

	//// datacontentencoding
	//// Type: String per RFC 2045 Section 6.1
	//// Constraints:
	////  The attribute MUST be set if the data attribute contains string-encoded binary data.
	////    Otherwise the attribute MUST NOT be set.
	////  If present, MUST adhere to RFC 2045 Section 6.1
	//if ec.DeprecatedDataContentEncoding != nil {
	//	dataContentEncoding := strings.ToLower(strings.TrimSpace(*ec.DeprecatedDataContentEncoding))
	//	if dataContentEncoding != Base64 {
	//		// TODO: need to test for RFC 2046
	//		errors = append(errors, "datacontentencoding: if present, MUST adhere to RFC 2045 Section 6.1")
	//	}
	//}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "\n"))
	}
	return nil
}

// String returns a pretty-printed representation of the EventContext.
func (ec EventContextV1) String() string {
	b := strings.Builder{}

	b.WriteString("Context Attributes,\n")

	b.WriteString("  specversion: " + ec.SpecVersion + "\n")
	b.WriteString("  type: " + ec.Type + "\n")
	b.WriteString("  source: " + ec.Source.String() + "\n")
	if ec.Subject != nil {
		b.WriteString("  subject: " + *ec.Subject + "\n")
	}
	b.WriteString("  id: " + ec.ID + "\n")
	if ec.Time != nil {
		b.WriteString("  time: " + ec.Time.String() + "\n")
	}
	if ec.DataSchema != nil {
		b.WriteString("  dataschema: " + ec.DataSchema.String() + "\n")
	}
	if ec.DataContentType != nil {
		b.WriteString("  datacontenttype: " + *ec.DataContentType + "\n")
	}
	//if ec.DeprecatedDataContentEncoding != nil {
	//	b.WriteString("  datacontentencoding: " + *ec.DeprecatedDataContentEncoding + "\n")
	//}

	if ec.Extensions != nil && len(ec.Extensions) > 0 {
		b.WriteString("Extensions,\n")
		keys := make([]string, 0, len(ec.Extensions))
		for k := range ec.Extensions {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			b.WriteString(fmt.Sprintf("  %s: %v\n", key, ec.Extensions[key]))
		}
	}

	return b.String()
}
