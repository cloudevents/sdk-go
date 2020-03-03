package event_test

import (
	"github.com/cloudevents/sdk-go/pkg/event"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStringOfApplicationJSON(t *testing.T) {
	want := strptr("application/json")
	got := event.StringOfApplicationJSON()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}

func TestStringOfApplicationXML(t *testing.T) {
	want := strptr("application/xml")
	got := event.StringOfApplicationXML()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}

func TestStringOfApplicationCloudEventsJSON(t *testing.T) {
	want := strptr("application/cloudevents+json")
	got := event.StringOfApplicationCloudEventsJSON()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}

func TestStringOfApplicationCloudEventsBatchJSON(t *testing.T) {
	want := strptr("application/cloudevents-batch+json")
	got := event.StringOfApplicationCloudEventsBatchJSON()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}
