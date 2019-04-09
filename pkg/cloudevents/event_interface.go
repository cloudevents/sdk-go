package cloudevents

import (
	"time"
)

type EventReader interface {
	// SpecVersion returns event.Context.GetSpecVersion().
	SpecVersion() string
	// Type returns event.Context.GetType().
	Type() string
	// Source returns event.Context.GetSource().
	Source() string
	// Subject returns event.Context.GetSubject().
	Subject() string
	// ID returns event.Context.GetID().
	ID() string
	// Time returns event.Context.GetTime().
	Time() time.Time
	// SchemaURL returns event.Context.GetSchemaURL().
	SchemaURL() string
	// DataContentType returns event.Context.GetDataContentType().
	DataContentType() string
	// DataMediaType returns event.Context.GetDataMediaType().
	DataMediaType() string
	// DataContentEncoding returns event.Context.GetDataContentEncoding().
	DataContentEncoding() string

	// Extension Attributes

	// ExtensionAs returns event.Context.ExtensionAs(name, obj).
	ExtensionAs(string, interface{}) error

	// Data Attribute

	// ExtensionAs returns event.Context.ExtensionAs(name, obj).
	DataAs(interface{}) error
}

type EventWriter interface {
	// Context Attributes

	// SetSpecVersion performs event.Context.SetSpecVersion.
	SetSpecVersion(string) error
	// SetType performs event.Context.SetType.
	SetType(string) error
	// SetSource performs event.Context.SetSource.
	SetSource(string) error
	// SetSubject( performs event.Context.SetSubject.
	SetSubject(string) error
	// SetID performs event.Context.SetID.
	SetID(string) error
	// SetTime performs event.Context.SetTime.
	SetTime(time.Time) error
	// SetSchemaURL performs event.Context.SetSchemaURL.
	SetSchemaURL(string) error
	// SetDataContentType performs event.Context.SetDataContentType.
	SetDataContentType(string) error
	// SetDataContentEncoding performs event.Context.SetDataContentEncoding.
	SetDataContentEncoding(string) error

	// Extension Attributes

	// SetExtension performs event.Context.SetExtension.
	SetExtension(string, interface{}) error

	// Data Attribute

	// SetData sets the data attribute.
	SetData(interface{}) error
}
