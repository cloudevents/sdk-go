package binding_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudevents/sdk-go/pkg/binding"
	cloudevents "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/types"
)

func TestWithFinish(t *testing.T) {
	var testEvent = event.Event{
		Data:        []byte(`"data"`),
		DataEncoded: true,
		Context: event.EventContextV1{
			DataContentType: cloudevents.StringOfApplicationJSON(),
			Source:          types.URIRef{URL: url.URL{Path: "source"}},
			ID:              "id",
			Type:            "type"}.AsV1(),
	}

	done := make(chan error, 1)
	m := binding.WithFinish(binding.EventMessage(testEvent), func(err error) {
		done <- err
	})
	select {
	case <-done:
		assert.Fail(t, "done early")
	default:
	}
	ch := make(chan binding.Message, 1)
	assert.NoError(t, binding.ChanSender(ch).Send(context.Background(), m))
	assert.NoError(t, <-done)
}
