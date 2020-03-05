package binding_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	client "github.com/cloudevents/sdk-go/pkg/client"
	"github.com/cloudevents/sdk-go/pkg/event"
)

func TestTransportSend(t *testing.T) {
	messageChannel := make(chan binding.Message, 1)
	transport := binding.NewTransportAdapter(binding.ChanSender(messageChannel), binding.ChanReceiver(messageChannel), nil)
	ev := test.MinEvent()

	c, err := client.NewWithTransport(transport)
	require.NoError(t, err)

	_, _, err = c.Send(context.Background(), ev)
	require.NoError(t, err)

	result := <-messageChannel

	test.AssertEventEquals(t, ev, event.Event(result.(binding.EventMessage)))
}

func TestTransportReceive(t *testing.T) {
	messageChannel := make(chan binding.Message, 1)
	eventReceivedChannel := make(chan event.Event, 1)
	transport := binding.NewTransportAdapter(binding.ChanSender(messageChannel), binding.ChanReceiver(messageChannel), nil)
	ev := test.MinEvent()

	c, err := client.NewWithTransport(transport)
	require.NoError(t, err)

	messageChannel <- binding.EventMessage(ev)

	go func() {
		err = c.StartReceiver(context.Background(), func(event event.Event) {
			eventReceivedChannel <- event
		})
		require.NoError(t, err)
	}()

	result := <-eventReceivedChannel

	test.AssertEventEquals(t, ev, result)
}
