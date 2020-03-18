package format_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudevents/sdk-go/v1/binding/format"
	ce "github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
)

func TestJSON(t *testing.T) {
	assert := assert.New(t)
	e := ce.Event{
		Context: ce.EventContextV03{
			Type:   "type",
			ID:     "id",
			Source: *types.ParseURLRef("source"),
		}.AsV03(),
	}
	e.SetExtension("ex", "val")
	assert.NoError(e.SetData("foo"))
	b, err := format.JSON.Marshal(e)
	assert.NoError(err)
	assert.Equal(`{"data":"foo","ex":"val","id":"id","source":"source","specversion":"0.3","type":"type"}`, string(b))

	var e2 ce.Event
	assert.NoError(format.JSON.Unmarshal(b, &e2))
	assert.Equal(e, e2)
}

func TestLookup(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(format.Lookup("nosuch"))

	f := format.Lookup(ce.ApplicationCloudEventsJSON)
	assert.Equal(f.MediaType(), ce.ApplicationCloudEventsJSON)
	assert.Equal(format.JSON, f)
}

func TestMarshalUnmarshal(t *testing.T) {
	assert := assert.New(t)
	e := ce.Event{
		Context: ce.EventContextV03{
			Type:   "type",
			ID:     "id",
			Source: *types.ParseURLRef("source"),
		}.AsV03(),
	}
	assert.NoError(e.SetData("foo"))
	b, err := format.Marshal(format.JSON.MediaType(), e)
	assert.NoError(err)
	assert.Equal(`{"data":"foo","id":"id","source":"source","specversion":"0.3","type":"type"}`, string(b))

	var e2 ce.Event
	assert.NoError(format.Unmarshal(format.JSON.MediaType(), b, &e2))
	assert.Equal(e, e2)

	_, err = format.Marshal("nosuchformat", e)
	assert.EqualError(err, "unknown event format media-type \"nosuchformat\"")
	err = format.Unmarshal("nosuchformat", nil, &e)
	assert.EqualError(err, "unknown event format media-type \"nosuchformat\"")
}

type dummyFormat struct{}

func (dummyFormat) MediaType() string                     { return "dummy" }
func (dummyFormat) Marshal(ce.Event) ([]byte, error)      { return []byte("dummy!"), nil }
func (dummyFormat) Unmarshal(b []byte, e *ce.Event) error { e.Data = "undummy!"; return nil }

func TestAdd(t *testing.T) {
	assert := assert.New(t)
	format.Add(dummyFormat{})
	assert.Equal(dummyFormat{}, format.Lookup("dummy"))

	e := ce.Event{}
	b, err := format.Marshal("dummy", e)
	assert.NoError(err)
	assert.Equal("dummy!", string(b))
	err = format.Unmarshal("dummy", b, &e)
	assert.NoError(err)
	assert.Equal("undummy!", e.Data)
}
