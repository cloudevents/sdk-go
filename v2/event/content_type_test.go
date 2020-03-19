package event_test

import (
	"testing"

	"github.com/cloudevents/sdk-go/v2/event"

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

func TestStringOfTextPlain(t *testing.T) {
	want := strptr("text/plain")
	got := event.StringOfTextPlain()

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
