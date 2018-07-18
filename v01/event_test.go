package v01_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/dispatchframework/cloudevents-go-sdk/v01"
)

func TestNewEvent(t *testing.T) {
	event := &v01.Event{
		EventType:        "testType",
		EventTypeVersion: "testVersion",
		Source:           "version",
		EventID:          "12345",
		EventTime:        nil,
		SchemaURL:        "http://example.com/schema",
		ContentType:      "application/json",
		Extensions:       nil,
		Data:             map[string]interface{}{"key": "value"},
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Errorf("JSON Error received: %v", err)
	}

	eventUnmarshaled := &v01.Event{}
	json.Unmarshal(data, eventUnmarshaled)
	if !reflect.DeepEqual(event, eventUnmarshaled) {
		t.Errorf("source event %#v and unmarshaled event %#v are not equal", event, eventUnmarshaled)
	}
}

func TestGetSet(t *testing.T) {
	event := &v01.Event{
		EventType:        "testType",
		EventTypeVersion: "testVersion",
		Source:           "version",
		EventID:          "12345",
		EventTime:        nil,
		SchemaURL:        "http://example.com/schema",
		ContentType:      "application/json",
		Extensions:       nil,
		Data:             map[string]interface{}{"key": "value"},
	}

	value, ok := event.Get("nonexistent")
	if ok {
		t.Error("Get ok for nonexistent key shoud be false, but isn't")
	}
	if value != nil {
		t.Error("Get value for nonexistent key should be nil, but isn't")
	}

	value, ok = event.Get("contentType")
	if !ok {
		t.Error("Get ok for existing key shoud be true, but isn't")

	}
	if value != "application/json" {
		t.Errorf("Get value for contentType should be application/json, but is %s", value)
	}

	event.Set("eventType", "newType")
	if event.EventType != "newType" {
		t.Errorf("expected eventType to be newType, got %s", event.EventType)
	}

	event.Set("ext", "somevalue")
	value, ok = event.Get("ext")
	if !ok {
		t.Error("Get ok for ext key shoud be false, but isn't")
	}
	if value != "somevalue" {
		t.Errorf("Get value for ext key should be somevalue, but is %s", value)
	}

}
