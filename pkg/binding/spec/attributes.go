package spec

import (
	"fmt"
	"time"

	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

// Kind is a version-independent identifier for a CloudEvent context attribute.
type Kind uint8

const (
	// Required cloudevents attributes
	ID Kind = iota
	Source
	SpecVersion
	Type
	// Optional cloudevents attributes
	DataContentType
	DataSchema
	Subject
	Time
)
const nAttrs = int(Time) + 1

var kindNames = [nAttrs]string{
	"id",
	"source",
	"specversion",
	"type",
	"datacontenttype",
	"dataschema",
	"subject",
	"time",
}

// String is a human-readable string, for a valid attribute name use Attribute.Name
func (k Kind) String() string { return kindNames[k] }

// IsRequired returns true for attributes defined as "required" by the CE spec.
func (k Kind) IsRequired() bool { return k < DataContentType }

// Attribute is a named attribute accessor.
// The attribute name is specific to a Version.
type Attribute interface {
	Kind() Kind
	// Name of the attribute with respect to the current spec Version()
	Name() string
	// Version of the spec that this attribute belongs to
	Version() Version
	// Get the value of this attribute from an event context
	Get(ce.EventContextReader) interface{}
	// Set the value of this attribute on an event context
	Set(ce.EventContextWriter, interface{}) error
}

// accessor provides Kind, Get, Set.
type accessor interface {
	Kind() Kind
	Get(ce.EventContextReader) interface{}
	Set(ce.EventContextWriter, interface{}) error
}

var acc = [nAttrs]accessor{
	&aStr{aKind(ID), ce.EventContextReader.GetID, ce.EventContextWriter.SetID},
	&aStr{aKind(Source), ce.EventContextReader.GetSource, ce.EventContextWriter.SetSource},
	&aStr{aKind(SpecVersion), ce.EventContextReader.GetSpecVersion, ce.EventContextWriter.SetSpecVersion},
	&aStr{aKind(Type), ce.EventContextReader.GetType, ce.EventContextWriter.SetType},
	&aStr{aKind(DataContentType), ce.EventContextReader.GetDataContentType, ce.EventContextWriter.SetDataContentType},
	&aStr{aKind(DataSchema), ce.EventContextReader.GetDataSchema, ce.EventContextWriter.SetDataSchema},
	&aStr{aKind(Subject), ce.EventContextReader.GetSubject, ce.EventContextWriter.SetSubject},
	&aTime{aKind(Time), ce.EventContextReader.GetTime, ce.EventContextWriter.SetTime},
}

// aKind implements Kind()
type aKind Kind

func (kind aKind) Kind() Kind { return Kind(kind) }

type aStr struct {
	aKind
	get func(ce.EventContextReader) string
	set func(ce.EventContextWriter, string) error
}

func (a *aStr) Get(c ce.EventContextReader) interface{} {
	if s := a.get(c); s != "" {
		return s
	}
	return nil // Treat blank as missing
}

func (a *aStr) Set(c ce.EventContextWriter, v interface{}) error {
	s, err := types.ToString(v)
	if err != nil {
		return fmt.Errorf("invalid value for %s: %#v", a.Kind(), v)
	}
	return a.set(c, s)
}

type aTime struct {
	aKind
	get func(ce.EventContextReader) time.Time
	set func(ce.EventContextWriter, time.Time) error
}

func (a *aTime) Get(c ce.EventContextReader) interface{} {
	if v := a.get(c); !v.IsZero() {
		return v
	}
	return nil // Treat zero time as missing.
}

func (a *aTime) Set(c ce.EventContextWriter, v interface{}) error {
	t, err := types.ToTime(v)
	if err != nil {
		return fmt.Errorf("invalid value for %s: %#v", a.Kind(), v)
	}
	return a.set(c, t)
}
