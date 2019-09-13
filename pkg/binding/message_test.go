package binding_test

import (
	"net/url"
	"testing"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/stretchr/testify/assert"
)

var testEvent = ce.Event{
	Data:    "data",
	Context: ce.EventContextV03{Source: types.URLRef{URL: url.URL{Path: "source"}}, ID: "id", Type: "type"}.AsV03(),
}

var testJSON = `{"data":"data","datacontenttype":"application/json","id":"id","source":"source","specversion":"0.3","type":"type"}`

func TestEventMessage(t *testing.T) {
	assert := assert.New(t)
	m, err := binding.BinaryEncoder{}.Encode(testEvent)
	assert.NoError(err)
	f, b := m.Structured()
	assert.Empty(f)
	assert.Empty(b)
	e, err := m.Event()
	assert.NoError(err)
	assert.Equal(testEvent, e)
}

func TestStructMessage(t *testing.T) {
	assert := assert.New(t)
	m, err := binding.StructEncoder{format.JSON}.Encode(testEvent)
	assert.NoError(err)
	f, b := m.Structured()
	assert.Equal(ce.ApplicationCloudEventsJSON, f)
	assert.Equal(testJSON, string(b))
	e, err := m.Event()
	assert.NoError(err)
	assert.Equal(testEvent.Context, e.Context)
	var s string
	assert.NoError(e.DataAs(&s))
	assert.Equal(testEvent.Data, s)

	_, err = binding.StructMessage{Format: "nosuch"}.Event()
	assert.EqualError(err, "unknown event format media-type \"nosuch\"")
}

type dummyFormat struct{}

func (dummyFormat) MediaType() string                     { return "dummy" }
func (dummyFormat) Marshal(ce.Event) ([]byte, error)      { return []byte("dummy!"), nil }
func (dummyFormat) Unmarshal(b []byte, e *ce.Event) error { e.Data = "undummy!"; return nil }

func TestStructured(t *testing.T) {
	sm := binding.StructMessage{Format: format.JSON.MediaType(), Bytes: []byte(testJSON)}
	b, err := binding.Structured(sm, format.JSON)
	assert.NoError(t, err)
	assert.Equal(t, &sm.Bytes, &b) // Not just equal but at the same address.

	d := dummyFormat{}
	b, err = binding.Structured(sm, d)
	assert.NoError(t, err)
	assert.Equal(t, "dummy!", string(b)) // Reformat as dummy

	bm := binding.EventMessage(testEvent)
	b, err = binding.Structured(bm, format.JSON)
	assert.NoError(t, err)
	assert.Equal(t, sm.Bytes, b) // Same bytes
}
