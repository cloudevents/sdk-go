package binding_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
	"github.com/cloudevents/sdk-go/v2/types"
)

func TestWithFinish(t *testing.T) {
	var testEvent = event.Event{
		DataEncoded: []byte(`"data"`),
		Context: event.EventContextV1{
			DataContentType: event.StringOfApplicationJSON(),
			Source:          types.URIRef{URL: url.URL{Path: "source"}},
			ID:              "id",
			Type:            "type"}.AsV1(),
	}

	done := make(chan error, 1)
	m := binding.WithFinish((*binding.EventMessage)(&testEvent), func(err error) {
		done <- err
	})
	select {
	case <-done:
		assert.Fail(t, "done early")
	default:
	}
	assert.NoError(t, m.Finish(nil))
	assert.NoError(t, <-done)
}

func TestUnwrap(t *testing.T) {
	testEvent := test.FullEvent()

	m := binding.WithFinish(binding.WithFinish(binding.ToMessage(&testEvent), func(err error) {}), func(err error) {})
	assert.Equal(t, binding.ToMessage(&testEvent), binding.UnwrapMessage(m))
}

func TestMessageMetadataHandler(t *testing.T) {
	var testEvent = test.FullEvent()
	finishMessage := binding.WithFinish((*binding.EventMessage)(&testEvent), func(err error) {})

	_, ty := finishMessage.(binding.MessageMetadataReader).GetAttribute(spec.Type)
	require.Equal(t, testEvent.Type(), ty)
	require.Equal(t, testEvent.Extensions()["exstring"], finishMessage.(binding.MessageMetadataReader).GetExtension("exstring"))
}
