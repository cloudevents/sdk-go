package test

import (
	"testing"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
)

func TestAssertEvent(t *testing.T) {
	tests := []struct {
		name       string
		have       event.Event
		assertions []EventMatcher
	}{{
		name:       "valid",
		have:       FullEvent(),
		assertions: []EventMatcher{IsValid()},
	}, {
		name:       "contains context attributes",
		have:       FullEvent(),
		assertions: []EventMatcher{IsValid(), ContainsAttributes(spec.ID, spec.SpecVersion)},
	}, {
		name:       "has exactly extensions",
		have:       FullEvent(),
		assertions: []EventMatcher{IsValid(), HasExactlyExtensions(FullEvent().Extensions())},
	}, {
		name:       "contains extensions",
		have:       FullEvent(),
		assertions: []EventMatcher{IsValid(), ContainsExtensions("exbool")},
	}, {
		name:       "has extension",
		have:       FullEvent(),
		assertions: []EventMatcher{IsValid(), HasExtension("exbool", true)},
	}, {
		name:       "has data",
		have:       FullEvent(),
		assertions: []EventMatcher{IsValid(), HasData(FullEvent().Data())},
	}, {
		name: "has attributes",
		have: FullEvent(),
		assertions: []EventMatcher{
			IsValid(),
			HasId(FullEvent().ID()),
			HasSource(FullEvent().Source()),
			HasSpecVersion(FullEvent().SpecVersion()),
			HasType(FullEvent().Type()),
			HasDataContentType(FullEvent().DataContentType()),
			HasDataSchema(FullEvent().DataSchema()),
			HasTime(FullEvent().Time()),
			HasSubject(FullEvent().Subject()),
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertEvent(t, tt.have, tt.assertions...)
		})
	}
}
