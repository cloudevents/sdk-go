package format_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/types"
)

func TestJSON(t *testing.T) {
	assert := assert.New(t)
	e := event.Event{
		Context: event.EventContextV03{
			Type:   "type",
			ID:     "id",
			Source: *types.ParseURIRef("source"),
		}.AsV03(),
	}
	e.SetExtension("ex", "val")
	assert.NoError(e.SetData("foo", event.ApplicationJSON))
	b, err := format.JSON.Marshal(&e)
	assert.NoError(err)
	assert.Equal(`{"data":"foo","datacontenttype":"application/json","ex":"val","id":"id","source":"source","specversion":"0.3","type":"type"}`, string(b))

	var e2 event.Event
	assert.NoError(format.JSON.Unmarshal(b, &e2))
	assert.Equal(e, e2)
}

func TestLookup(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(format.Lookup("nosuch"))

	f := format.Lookup(event.ApplicationCloudEventsJSON)
	assert.Equal(f.MediaType(), event.ApplicationCloudEventsJSON)
	assert.Equal(format.JSON, f)
}

func TestMarshalUnmarshal(t *testing.T) {
	assert := assert.New(t)
	e := event.Event{
		Context: event.EventContextV03{
			Type:   "type",
			ID:     "id",
			Source: *types.ParseURIRef("source"),
		}.AsV03(),
	}
	assert.NoError(e.SetData("foo", event.ApplicationJSON))
	b, err := format.Marshal(format.JSON.MediaType(), &e)
	assert.NoError(err)
	assert.Equal(`{"data":"foo","datacontenttype":"application/json","id":"id","source":"source","specversion":"0.3","type":"type"}`, string(b))

	var e2 event.Event
	assert.NoError(format.Unmarshal(format.JSON.MediaType(), b, &e2))
	assert.Equal(e, e2)

	_, err = format.Marshal("nosuchformat", &e)
	assert.EqualError(err, "unknown event format media-type \"nosuchformat\"")
	err = format.Unmarshal("nosuchformat", nil, &e)
	assert.EqualError(err, "unknown event format media-type \"nosuchformat\"")
}

type dummyFormat struct{}

func (dummyFormat) MediaType() string                    { return "dummy" }
func (dummyFormat) Marshal(*event.Event) ([]byte, error) { return []byte("dummy!"), nil }
func (dummyFormat) Unmarshal(b []byte, e *event.Event) error {
	e.DataEncoded = []byte("undummy!")
	return nil
}

func TestAdd(t *testing.T) {
	assert := assert.New(t)
	format.Add(dummyFormat{})
	assert.Equal(dummyFormat{}, format.Lookup("dummy"))

	e := event.Event{}
	b, err := format.Marshal("dummy", &e)
	assert.NoError(err)
	assert.Equal("dummy!", string(b))
	err = format.Unmarshal("dummy", b, &e)
	assert.NoError(err)
	assert.Equal([]byte("undummy!"), e.Data())
}
