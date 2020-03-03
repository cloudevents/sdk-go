package event_test

import (
	"github.com/cloudevents/sdk-go/pkg/event"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStringOfBase64(t *testing.T) {
	want := strptr("base64")
	got := event.StringOfBase64()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}
