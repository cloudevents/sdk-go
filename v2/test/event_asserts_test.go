package test

import (
	"testing"

	"github.com/stretchr/testify/require"

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

func TestAssertAnyOf(t *testing.T) {
	tests := []struct {
		name      string
		have      event.Event
		anyOf     []EventMatcher
		shouldErr bool
	}{{
		name:      "any of the two ids matching",
		have:      FullEvent(),
		anyOf:     []EventMatcher{HasId("min-event"), HasId("full-event")},
		shouldErr: false,
	}, {
		name:      "any of the two ids matching - reverse",
		have:      FullEvent(),
		anyOf:     []EventMatcher{HasId("full-event"), HasId("min-event")},
		shouldErr: false,
	}, {
		name:      "none matching",
		have:      FullEvent(),
		anyOf:     []EventMatcher{HasId("other-event"), HasId("min-event")},
		shouldErr: true,
	}, {
		name:      "both matching",
		have:      FullEvent(),
		anyOf:     []EventMatcher{HasId("full-event"), HasId("full-event")},
		shouldErr: false,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldErr {
				require.NoError(t, AnyOf(tt.anyOf...)(tt.have))
			} else {
				require.NoError(t, AnyOf(tt.anyOf...)(tt.have))
			}
		})
	}
}
