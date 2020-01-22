package binding

import (
	"context"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

func TestWithAcksBeforeFinish(t *testing.T) {
	var testEvent = cloudevents.Event{
		Data:        []byte(`"data"`),
		DataEncoded: true,
		Context: cloudevents.EventContextV1{
			DataContentType: cloudevents.StringOfApplicationJSON(),
			Source:          types.URIRef{URL: url.URL{Path: "source"}},
			ID:              "id",
			Type:            "type"}.AsV1(),
	}

	finishCalled := false
	finishMessage := WithFinish(EventMessage(testEvent), func(err error) {
		finishCalled = true
	})

	wg := sync.WaitGroup{}

	messageToTest := WithAcksBeforeFinish(finishMessage, 1000)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(m Message) {
			ch := make(chan Message, 1)
			assert.NoError(t, ChanSender(ch).Send(context.Background(), m))
			<-ch
			wg.Done()
		}(messageToTest)
	}

	wg.Wait()
	assert.True(t, finishCalled)
}

func TestCopyAndWithAcksBeforeFinish(t *testing.T) {
	var testEvent = cloudevents.Event{
		Data:        []byte(`"data"`),
		DataEncoded: true,
		Context: cloudevents.EventContextV1{
			DataContentType: cloudevents.StringOfApplicationJSON(),
			Source:          types.URIRef{URL: url.URL{Path: "source"}},
			ID:              "id",
			Type:            "type"}.AsV1(),
	}

	finishCalled := false
	finishMessage := WithFinish(EventMessage(testEvent), func(err error) {
		finishCalled = true
	})

	copiedMessage, err := CopyMessage(finishMessage)
	assert.NoError(t, err)

	wg := sync.WaitGroup{}

	messageToTest := WithAcksBeforeFinish(copiedMessage, 1000)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(m Message) {
			ch := make(chan Message, 1)
			assert.NoError(t, ChanSender(ch).Send(context.Background(), m))
			<-ch
			wg.Done()
		}(messageToTest)
	}

	wg.Wait()
	assert.True(t, finishCalled)
}
