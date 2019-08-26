package cloudevents_test

import (
	"testing"

	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/google/go-cmp/cmp"
)

func TestStringOfBase64(t *testing.T) {
	want := strptr("base64")
	got := ce.StringOfBase64()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected string (-want, +got) = %v", diff)
	}
}
