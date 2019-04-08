package cloudevents

import (
	"time"
)

type EventReader interface {
	// Context Attributes

	SpecVersion() string
	Type() string
	Source() string
	ID() string
	Time() time.Time
	SchemaURL() string
	DataContentType() string
	DataContentEncoding() string

	// Extension Attributes

	ExtensionAs(string, interface{}) error

	// Data Attribute

	DataAs(interface{}) error
}

type EventWriter interface {
	// Context Attributes

	SetSpecVersion(string) error
	SetType(string) error
	SetSource(string) error
	SetID(string) error
	SetTime(time time.Time) error
	SetSchemaURL(string) error
	SetDataContentType(string) error
	SetDataContentEncoding(string) error

	// Extension Attributes

	SetExtension(string, interface{}) error

	// Data Attribute

	SetData(data interface{}) error
}
