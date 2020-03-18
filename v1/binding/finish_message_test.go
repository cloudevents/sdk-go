package binding_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	cloudevents "github.com/cloudevents/sdk-go/v1"
	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
)

func TestWithFinish(t *testing.T) {
	var testEvent = cloudevents.Event{
		Data:        []byte(`"data"`),
		DataEncoded: true,
		Context: cloudevents.EventContextV1{
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
