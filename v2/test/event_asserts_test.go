package test

import (
	"testing"

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
		assertions: []EventMatcher{IsValid(), ContainsAttributes("id", "specversion")},
	}, {
		name: "has context attributes",
		have: FullEvent(),
		assertions: []EventMatcher{IsValid(), HasAttributes(map[string]interface{}{
			"id":          FullEvent().ID(),
			"specversion": event.CloudEventsVersionV1,
		})},
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
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AssertEvent(t, tt.have, tt.assertions...)
		})
	}
}
