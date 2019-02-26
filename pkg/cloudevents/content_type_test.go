package cloudevents_test

import (
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestStringOfApplicationJSON(t *testing.T) {
	want := strptr("application/json")
	got := ce.StringOfApplicationJSON()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}

func TestStringOfApplicationXML(t *testing.T) {
	want := strptr("application/xml")
	got := ce.StringOfApplicationXML()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}

func TestStringOfApplicationCloudEventsJSON(t *testing.T) {
	want := strptr("application/cloudevents+json")
	got := ce.StringOfApplicationCloudEventsJSON()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}

func TestStringOfApplicationCloudEventsBatchJSON(t *testing.T) {
	want := strptr("application/cloudevents-batch+json")
	got := ce.StringOfApplicationCloudEventsBatchJSON()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}
