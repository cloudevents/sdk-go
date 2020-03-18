package binding_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cloudevents "github.com/cloudevents/sdk-go/v1"
	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/test"
)

func TestTransportSend(t *testing.T) {
	messageChannel := make(chan binding.Message, 1)
	transport := binding.NewTransportAdapter(binding.ChanSender(messageChannel), binding.ChanReceiver(messageChannel), nil)
	ev := test.MinEvent()

	client, err := cloudevents.NewClient(transport, cloudevents.WithoutTracePropagation())
	require.NoError(t, err)

	_, _, err = client.Send(context.Background(), ev)
	require.NoError(t, err)

	result := <-messageChannel

	test.AssertEventEquals(t, ev, cloudevents.Event(result.(binding.EventMessage)))
}

func TestTransportReceive(t *testing.T) {
	messageChannel := make(chan binding.Message, 1)
	eventReceivedChannel := make(chan cloudevents.Event, 1)
	transport := binding.NewTransportAdapter(binding.ChanSender(messageChannel), binding.ChanReceiver(messageChannel), nil)
	ev := test.MinEvent()

	client, err := cloudevents.NewClient(transport)
	require.NoError(t, err)

	messageChannel <- binding.EventMessage(ev)

	go func() {
		err = client.StartReceiver(context.Background(), func(event cloudevents.Event) {
			eventReceivedChannel <- event
		})
		require.NoError(t, err)
	}()

	result := <-eventReceivedChannel

	test.AssertEventEquals(t, ev, result)
}
