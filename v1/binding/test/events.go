// Package test contains test data and generic tests for testing bindings.
package test

import (
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/cloudevents/sdk-go/v1/binding/spec"
	ce "github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
)

func strptr(s string) *string { return &s }

var (
	Source    = types.URIRef{URL: url.URL{Scheme: "http", Host: "example.com", Path: "/source"}}
	Timestamp = types.Timestamp{Time: time.Date(2020, 03, 21, 12, 34, 56, 780000000, time.UTC)}
	Schema    = types.URI{URL: url.URL{Scheme: "http", Host: "example.com", Path: "/schema"}}
)

// FullEvent has all context attributes set and JSON string data.
func FullEvent() ce.Event {
	e := ce.Event{
		Context: ce.EventContextV1{
			Type:            "com.example.FullEvent",
			Source:          Source,
			ID:              "full-event",
			Time:            &Timestamp,
			DataSchema:      &Schema,
			DataContentType: strptr("text/json"),
			Subject:         strptr("topic"),
		}.AsV1(),
	}

	e.SetExtension("exbool", true)
	e.SetExtension("exint", 42)
	e.SetExtension("exstring", "exstring")
	e.SetExtension("exbinary", []byte{0, 1, 2, 3})
	e.SetExtension("exurl", Source)
	e.SetExtension("extime", Timestamp)

	if err := e.SetData("hello"); err != nil {
		panic(err)
	}
	return e
}

// MinEvent has only required attributes set.
func MinEvent() ce.Event {
	return ce.Event{
		Context: ce.EventContextV1{
			Type:   "com.example.MinEvent",
			Source: Source,
			ID:     "min-event",
		}.AsV1(),
	}
}

// AllVersions returns all versions of each event in events.
// ID gets a -number suffix so IDs are unique.
func AllVersions(events []ce.Event) []ce.Event {
	versions := spec.New()
	all := versions.Versions()
	result := make([]ce.Event, len(events)*len(all))
	i := 0
	for _, e := range events {
		for _, v := range all {
			result[i] = e
			result[i].Context = v.Convert(e.Context)
			result[i].SetID(fmt.Sprintf("%v-%v", e.ID(), i)) // Unique IDs
			i++
		}
	}
	return result
}

// Events is a set of test events that should be handled correctly by
// all event-processing code.
func Events() []ce.Event {
	return AllVersions([]ce.Event{FullEvent(), MinEvent()})
}

// NoExtensions returns a copy of events with no Extensions.
// Use for testing where extensions are not supported.
func NoExtensions(events []ce.Event) []ce.Event {
	result := make([]ce.Event, len(events))
	for i, e := range events {
		result[i] = e
		result[i].Context = e.Context.Clone()
		ctx := reflect.ValueOf(result[i].Context).Elem()
		ext := ctx.FieldByName("Extensions")
		ext.Set(reflect.Zero(ext.Type()))
	}
	return result
}
