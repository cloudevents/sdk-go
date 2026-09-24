package event_test

import (
	"encoding/json"
	"testing"

	"github.com/cloudevents/sdk-go/v2/event"
)

func TestZeroLengthSliceMarshalJSON(t *testing.T) {
	ev := event.New(event.CloudEventsVersionV1)
	ev.SetDataContentType(event.ApplicationJSON)
	ev.SetType("type")
	ev.SetSource("source")
	ev.SetID("id")
	ev.DataEncoded = []byte{}

	if err := ev.Validate(); err != nil {
		t.Fatalf("validate failed: %s", err)
	}

	dta, err := ev.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal failed: %s", err)
	}

	// Check if it's valid JSON
	var raw json.RawMessage
	if err := json.Unmarshal(dta, &raw); err != nil {
		t.Fatalf("output is NOT valid JSON: %s\noutput: %s", err, string(dta))
	}
	t.Logf("JSON output: %s", string(dta))

	var res event.Event
	if err := res.UnmarshalJSON(dta); err != nil {
		t.Fatalf("unmarshal failed: %s", err)
	}
}
