package v01_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvent(t *testing.T) {
	timestamp, err := time.Parse(time.RFC3339, "2018-04-05T17:31:00Z")
	require.NoError(t, err)

	event := &v01.Event{
		EventType:        "com.example.someevent",
		EventTypeVersion: "1.1",
		Source:           "/mycontext",
		EventID:          "1234-1234-1234",
		EventTime:        &timestamp,
		SchemaURL:        "http://example.com/schema",
		ContentType:      "application/json",
		Data:             map[string]interface{}{"key": "value"},
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Errorf("JSON Error received: %v", err)
	}
	fmt.Printf("%s", data)

	eventUnmarshaled := &v01.Event{}
	json.Unmarshal(data, eventUnmarshaled)
	assert.EqualValues(t, event, eventUnmarshaled)
}

func TestGetSet(t *testing.T) {
	event := &v01.Event{
		EventType:        "com.example.someevent",
		EventTypeVersion: "1.1",
		Source:           "/mycontext",
		EventID:          "1234-1234-1234",
		EventTime:        nil,
		SchemaURL:        "http://example.com/schema",
		ContentType:      "application/json",
		Data:             map[string]interface{}{"key": "value"},
	}

	value, ok := event.Get("nonexistent")
	assert.False(t, ok, "ok should be false for nonexistent key, but isn't")
	assert.Nil(t, value, "value for nonexistent key should be nil, but isn't")

	value, ok = event.Get("contentType")
	assert.True(t, ok, "ok for existing key should be true, but isn't")
	assert.Equal(t, "application/json", value, "value for contentType should be application/json, but is %s", value)

	event.Set("eventType", "newType")
	assert.Equal(t, "newType", event.EventType, "expected eventType to be 'newType', got %s", event.EventType)

	event.Set("ext", "somevalue")
	value, ok = event.Get("ext")
	assert.True(t, ok, "ok for ext key should be true, but isn't")
	assert.Equal(t, "somevalue", value, "value for ext key should be 'somevalue', but is %s", value)
}

func TestProperties(t *testing.T) {
	event := v01.Event{}

	props := event.Properties()

	assert.True(t, props["eventid"])
	delete(props, "eventid")
	assert.True(t, props["source"])
	delete(props, "source")
	assert.True(t, props["eventtype"])
	delete(props, "eventtype")
	assert.True(t, props["cloudeventsversion"])
	delete(props, "cloudeventsversion")

	for k, v := range props {
		assert.False(t, v, "property %s should not be required.", k)
	}
}

func TestUnmarshallJSON(t *testing.T) {

	var actual v01.Event
	err := json.Unmarshal([]byte("{\"eventType\":\"com.example.someevent\", \"eventTime\":\"2018-04-05T17:31:00Z\", \"myextension\":\"myValue\", \"data\": {\"topKey\" : \"topValue\", \"objectKey\": {\"embedKey\" : \"embedValue\"} }}"), &actual)
	assert.NoError(t, err)

	timestamp, _ := time.Parse(time.RFC3339, "2018-04-05T17:31:00Z")
	expected := v01.Event{
		EventType: "com.example.someevent",
		EventTime: &timestamp,
		Data: map[string]interface{}{
			"topKey": "topValue",
			"objectKey": map[string]interface{}{
				"embedKey": "embedValue",
			},
		},
	}

	expected.Set("myExtension", "myValue")
	assert.EqualValues(t, expected, actual)
}

func TestMarshallJSON(t *testing.T) {
	timestamp, _ := time.Parse(time.RFC3339, "2018-04-05T17:31:00Z")
	input := v01.Event{
		CloudEventsVersion: "0.1",
		EventID:            "1234-1234-1234",
		EventType:          "com.example.someevent",
		Source:             "/mycontext",
		EventTime:          &timestamp,
		Data: map[string]interface{}{
			"topKey": "topValue",
			"objectKey": map[string]interface{}{
				"embedKey": "embedValue",
			},
		},
	}
	input.Set("myExtension", "myValue")

	actual, err := json.Marshal(input)
	expected := []byte("{\"cloudEventsVersion\":\"0.1\",\"data\":{\"objectKey\":{\"embedKey\":\"embedValue\"},\"topKey\":\"topValue\"},\"eventID\":\"1234-1234-1234\",\"eventTime\":\"2018-04-05T17:31:00Z\",\"eventType\":\"com.example.someevent\",\"myextension\":\"myValue\",\"source\":\"/mycontext\"}")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
