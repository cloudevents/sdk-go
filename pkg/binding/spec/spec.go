package spec

import (
	"fmt"
	"strings"

	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

// Version provides meta-data for a single spec-version.
type Version interface {
	// String name of the version, e.g. "1.0"
	String() string
	// Prefix for attribute names.
	Prefix() string
	// Attribute looks up a prefixed attribute name (case insensitive).
	// Returns nil if not found.
	Attribute(name string) Attribute
	// Attributes returns all the context attributes for this version.
	Attributes() []Attribute
	// NewContext returns a new context for this version.
	NewContext() ce.EventContext
	// Convert translates a context to this version.
	Convert(ce.EventContextConverter) ce.EventContext
	// SetAttribute sets named attribute to value.
	//
	// Name is case insensitive.
	// Does nothing if name does not start with prefix.
	SetAttribute(context ce.EventContextWriter, name string, value interface{}) error
}

// Versions contains all known versions with the same attribute prefix.
type Versions struct {
	prefix  string
	all     []Version
	m       map[string]Version
	svnames []string
}

// Versions returns the list of all known versions, most recent first.
func (vs *Versions) Versions() []Version { return vs.all }

// Version returns the named version.
func (vs *Versions) Version(name string) (Version, error) {
	if v := vs.m[name]; v != nil {
		return v, nil
	}
	return nil, fmt.Errorf("invalid spec version %#v", name)
}

// Latest returns the latest Version
func (vs *Versions) Latest() Version { return vs.all[0] }

// SpecVersionNames returns distinct names of the specversion
// attribute used in all versions, newest first.
// Names are prefixed.
func (vs *Versions) SpecVersionNames() []string { return vs.svnames }

// Prefix is the lowercase attribute name prefix.
func (vs *Versions) Prefix() string { return vs.prefix }

// FindVersion calls getAttr with known (prefixed) spec-version attribute names
// till it finds a valid version.
func (vs *Versions) FindVersion(getAttr func(string) string) (Version, error) {
	for _, sv := range vs.svnames {
		if v, err := vs.Version(getAttr(sv)); err == nil {
			return v, nil
		}
	}
	return nil, fmt.Errorf("CloudEvents spec-version not found")
}

type attribute struct {
	accessor
	name    string
	version Version
}

func (a *attribute) Name() string     { return a.name }
func (a *attribute) Version() Version { return a.version }

type version struct {
	prefix  string
	context ce.EventContext
	convert func(ce.EventContextConverter) ce.EventContext
	attrMap map[string]Attribute
	attrs   []Attribute
}

func (v *version) Attribute(name string) Attribute { return v.attrMap[strings.ToLower(name)] }
func (v *version) Attributes() []Attribute         { return v.attrs }
func (v *version) String() string                  { return v.context.GetSpecVersion() }
func (v *version) Prefix() string                  { return v.prefix }
func (v *version) NewContext() ce.EventContext     { return v.context.Clone() }

// HasPrefix is a case-insensitive prefix check.
func (v *version) HasPrefix(name string) bool {
	return strings.HasPrefix(strings.ToLower(name), v.prefix)
}

func (v *version) Convert(c ce.EventContextConverter) ce.EventContext { return v.convert(c) }

func (v *version) SetAttribute(c ce.EventContextWriter, name string, value interface{}) error {
	if a := v.Attribute(name); a != nil { // Standard attribute
		return a.Set(c, value)
	}
	name = strings.ToLower(name)
	var err error
	if strings.HasPrefix(name, v.prefix) { // Extension attribute
		value, err = types.Validate(value)
		if err == nil {
			err = c.SetExtension(strings.TrimPrefix(name, v.prefix), value)
		}
	}
	return err
}

func newVersion(
	prefix string,
	context ce.EventContext,
	convert func(ce.EventContextConverter) ce.EventContext,
	attrs ...*attribute,
) *version {
	v := &version{
		prefix:  strings.ToLower(prefix),
		context: context,
		convert: convert,
		attrMap: map[string]Attribute{},
		attrs:   make([]Attribute, len(attrs)),
	}
	for i, a := range attrs {
		a.name = prefix + a.name
		a.version = v
		v.attrs[i] = a
		v.attrMap[strings.ToLower(a.name)] = a
	}
	return v
}

// WithPrefix returns a set of versions with prefix added to all attribute names.
func WithPrefix(prefix string) *Versions {
	attr := func(name string, kind Kind) *attribute {
		return &attribute{accessor: acc[kind], name: name}
	}
	vs := &Versions{
		m: map[string]Version{},
		svnames: []string{
			prefix + "specversion",
			prefix + "cloudEventsVersion",
		},
		all: []Version{
			newVersion(prefix, ce.EventContextV1{}.AsV1(),
				func(c ce.EventContextConverter) ce.EventContext { return c.AsV1() },
				attr("id", ID),
				attr("source", Source),
				attr("specversion", SpecVersion),
				attr("type", Type),
				attr("datacontenttype", DataContentType),
				attr("dataschema", DataSchema),
				attr("subject", Subject),
				attr("time", Time),
			),
			newVersion(prefix, ce.EventContextV03{}.AsV03(),
				func(c ce.EventContextConverter) ce.EventContext { return c.AsV03() },
				attr("specversion", SpecVersion),
				attr("type", Type),
				attr("source", Source),
				attr("schemaurl", DataSchema),
				attr("subject", Subject),
				attr("id", ID),
				attr("time", Time),
				attr("datacontenttype", DataContentType),
			),
			newVersion(prefix, ce.EventContextV02{}.AsV02(),
				func(c ce.EventContextConverter) ce.EventContext { return c.AsV02() },
				attr("specversion", SpecVersion),
				attr("type", Type),
				attr("source", Source),
				attr("schemaurl", DataSchema),
				attr("id", ID),
				attr("time", Time),
				attr("contenttype", DataContentType),
			),
			newVersion(prefix, ce.EventContextV01{}.AsV01(),
				func(c ce.EventContextConverter) ce.EventContext { return c.AsV01() },
				attr("cloudEventsVersion", SpecVersion),
				attr("eventType", Type),
				attr("source", Source),
				attr("schemaURL", DataSchema),
				attr("eventID", ID),
				attr("eventTime", Time),
				attr("contentType", DataContentType),
			),
		},
	}
	for _, v := range vs.all {
		vs.m[v.String()] = v
	}
	return vs
}

// New returns a set of versions
func New() *Versions { return WithPrefix("") }

// Built-in un-prefixed versions.
var (
	V01 Version
	V02 Version
	V03 Version
	V1  Version
)

func init() {
	vs := New()
	V01, _ = vs.Version("0.1")
	V02, _ = vs.Version("0.2")
	V03, _ = vs.Version("0.3")
	V1, _ = vs.Version("1.0")
}
