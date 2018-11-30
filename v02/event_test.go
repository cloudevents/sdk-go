package v02_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v02"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvent(t *testing.T) {
	timestamp, err := time.Parse(time.RFC3339, "2018-04-05T17:31:00Z")
	require.NoError(t, err)

	event := &v02.Event{
		Type:        "com.example.someevent",
		Source:      "/mycontext",
		ID:          "1234-1234-1234",
		Time:        &timestamp,
		SchemaURL:   "http://example.com/schema",
		ContentType: "application/json",
		Data:        map[string]interface{}{"key": "value"},
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Errorf("JSON Error received: %v", err)
	}
	fmt.Printf("%s", data)

	eventUnmarshaled := &v02.Event{}
	json.Unmarshal(data, eventUnmarshaled)
	assert.EqualValues(t, event, eventUnmarshaled)
}

func TestGetSet(t *testing.T) {
	event := &v02.Event{
		Type:        "com.example.someevent",
		Source:      "/mycontext",
		ID:          "1234-1234-1234",
		Time:        nil,
		SchemaURL:   "http://example.com/schema",
		ContentType: "application/json",
		Data:        map[string]interface{}{"key": "value"},
	}

	value, ok := event.Get("nonexistent")
	assert.False(t, ok, "ok should be false for nonexistent key, but isn't")
	assert.Nil(t, value, "value for nonexistent key should be nil, but isn't")

	value, ok = event.Get("contentType")
	assert.True(t, ok, "ok for existing key should be true, but isn't")
	assert.Equal(t, "application/json", value, "value for contentType should be application/json, but is %s", value)

	event.Set("type", "newType")
	assert.Equal(t, "newType", event.Type, "expected type to be 'newType', got %s", event.Type)

	event.Set("ext", "somevalue")
	value, ok = event.Get("ext")
	assert.True(t, ok, "ok for ext key should be true, but isn't")
	assert.Equal(t, "somevalue", value, "value for ext key should be 'somevalue', but is %s", value)
}

func TestProperties(t *testing.T) {
	event := v02.Event{}

	props := event.Properties()

	assert.True(t, props["id"])
	delete(props, "id")
	assert.True(t, props["source"])
	delete(props, "source")
	assert.True(t, props["type"])
	delete(props, "type")
	assert.True(t, props["specversion"])
	delete(props, "specversion")

	for k, v := range props {
		assert.False(t, v, "property %s should not be required.", k)
	}
}

func TestUnmarshallJSON(t *testing.T) {

	var actual v02.Event
	err := json.Unmarshal([]byte("{\"type\":\"com.example.someevent\", \"time\":\"2018-04-05T17:31:00Z\", \"myextension\":\"myValue\", \"data\": {\"topKey\" : \"topValue\", \"objectKey\": {\"embedKey\" : \"embedValue\"} }}"), &actual)
	assert.NoError(t, err)

	timestamp, _ := time.Parse(time.RFC3339, "2018-04-05T17:31:00Z")
	expected := v02.Event{
		Type: "com.example.someevent",
		Time: &timestamp,
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
	input := v02.Event{
		SpecVersion: "0.2",
		ID:          "1234-1234-1234",
		Type:        "com.example.someevent",
		Source:      "/mycontext",
		Time:        &timestamp,
		Data: map[string]interface{}{
			"topKey": "topValue",
			"objectKey": map[string]interface{}{
				"embedKey": "embedValue",
			},
		},
	}
	input.Set("myExtension", "myValue")

	actualBytes, err := json.Marshal(input)
	assert.NoError(t, err)

	var output v02.Event
	err = json.Unmarshal(actualBytes, &output)
	assert.NoError(t, err)
	assert.Equal(t, input, output)
}
