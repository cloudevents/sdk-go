package buffering

import (
	"context"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
	"github.com/cloudevents/sdk-go/v2/types"
)

func TestWithAcksBeforeFinish(t *testing.T) {
	var testEvent = event.Event{
		DataEncoded: []byte(`"data"`),
		Context: event.EventContextV1{
			DataContentType: event.StringOfApplicationJSON(),
			Source:          types.URIRef{URL: url.URL{Path: "source"}},
			ID:              "id",
			Type:            "type"}.AsV1(),
	}

	finishCalled := false
	finishMessage := binding.WithFinish((*binding.EventMessage)(&testEvent), func(err error) {
		finishCalled = true
	})

	wg := sync.WaitGroup{}

	messageToTest := WithAcksBeforeFinish(finishMessage, 1000)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(m binding.Message) {
			require.False(t, finishCalled)
			require.NoError(t, messageToTest.Finish(nil))
			wg.Done()
		}(messageToTest)
	}

	wg.Wait()
	require.True(t, finishCalled)
}

func TestCopyAndWithAcksBeforeFinish(t *testing.T) {
	var testEvent = event.Event{
		DataEncoded: []byte(`"data"`),
		Context: event.EventContextV1{
			DataContentType: event.StringOfApplicationJSON(),
			Source:          types.URIRef{URL: url.URL{Path: "source"}},
			ID:              "id",
			Type:            "type"}.AsV1(),
	}

	finishCalled := false
	finishMessage := binding.WithFinish((*binding.EventMessage)(&testEvent), func(err error) {
		finishCalled = true
	})

	copiedMessage, err := BufferMessage(context.Background(), finishMessage)
	require.NoError(t, err)

	wg := sync.WaitGroup{}

	messageToTest := WithAcksBeforeFinish(copiedMessage, 1000)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(m binding.Message) {
			require.False(t, finishCalled)
			require.NoError(t, messageToTest.Finish(nil))
			wg.Done()
		}(messageToTest)
	}

	wg.Wait()
	require.True(t, finishCalled)
}

func TestMessageMetadataHandler(t *testing.T) {
	var testEvent = test.FullEvent()
	finishMessage := WithAcksBeforeFinish((*binding.EventMessage)(&testEvent), 3)

	_, ty := finishMessage.(binding.MessageMetadataReader).GetAttribute(spec.Type)
	require.Equal(t, testEvent.Type(), ty)
	require.Equal(t, testEvent.Extensions()["exstring"], finishMessage.(binding.MessageMetadataReader).GetExtension("exstring"))
}
