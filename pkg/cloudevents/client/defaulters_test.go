package client

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/go-cmp/cmp"
	"testing"
	"time"
)

func TestDefaultIDToUUIDIfNotSet(t *testing.T) {
	testCases := map[string]struct {
		event cloudevents.Event
	}{
		"nil context": {
			event: cloudevents.Event{},
		},
		"v0.1 empty": {
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV01{},
			},
		},
		"v0.2 empty": {
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV02{},
			},
		},
		"v0.3 empty": {
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{},
			},
		},
		"v0.1 no change": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV01{EventID: "abc"}.AsV01(),
			},
		},
		"v0.2 no change": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV02{ID: "abc"}.AsV02(),
			},
		},
		"v0.3 no change": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV03{ID: "abc"}.AsV03(),
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := DefaultIDToUUIDIfNotSet(tc.event)

			if got.Context != nil && got.Context.AsV02().ID == "" {
				t.Errorf("failed to generate an id for event")
			}
		})
	}
}

func TestDefaultIDToUUIDIfNotSetImmutable(t *testing.T) {
	event := cloudevents.Event{
		Context: &cloudevents.EventContextV01{},
	}

	got := DefaultIDToUUIDIfNotSet(event)

	want := "0.1"

	if diff := cmp.Diff(want, got.SpecVersion()); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}

	if event.Context.AsV01().EventID != "" {
		t.Errorf("modified the original event")
	}

	if got.Context.AsV01().EventID == "" {
		t.Errorf("failed to generate an id for event")
	}
}

func TestDefaultTimeToNowIfNotSet(t *testing.T) {
	testCases := map[string]struct {
		event cloudevents.Event
	}{
		"nil context": {
			event: cloudevents.Event{},
		},
		"v0.1 empty": {
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV01{},
			},
		},
		"v0.2 empty": {
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV02{},
			},
		},
		"v0.3 empty": {
			event: cloudevents.Event{
				Context: &cloudevents.EventContextV03{},
			},
		},
		"v0.1 no change": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV01{EventTime: &types.Timestamp{Time: time.Now()}}.AsV01(),
			},
		},
		"v0.2 no change": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV02{Time: &types.Timestamp{Time: time.Now()}}.AsV02(),
			},
		},
		"v0.3 no change": {
			event: cloudevents.Event{
				Context: cloudevents.EventContextV03{Time: &types.Timestamp{Time: time.Now()}}.AsV03(),
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			got := DefaultTimeToNowIfNotSet(tc.event)

			if got.Context != nil && got.Context.AsV02().Time.IsZero() {
				t.Errorf("failed to generate time for event")
			}
		})
	}
}

func TestDefaultTimeToNowIfNotSetImmutable(t *testing.T) {
	event := cloudevents.Event{
		Context: &cloudevents.EventContextV01{},
	}

	got := DefaultTimeToNowIfNotSet(event)

	want := "0.1"

	if diff := cmp.Diff(want, got.SpecVersion()); diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}

	if event.Context.AsV01().EventTime != nil {
		t.Errorf("modified the original event")
	}

	if got.Context.AsV01().EventTime.IsZero() {
		t.Errorf("failed to generate a time for event")
	}
}
