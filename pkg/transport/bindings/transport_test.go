package bindings_test

import (
	"context"
	"testing"

	bindings2 "github.com/cloudevents/sdk-go/pkg/transport/bindings"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	client "github.com/cloudevents/sdk-go/pkg/client"
	"github.com/cloudevents/sdk-go/pkg/event"
)

func TestTransportSend(t *testing.T) {
	messageChannel := make(chan binding.Message, 1)
	transport := bindings2.NewSendingTransport(binding.ChanSender(messageChannel), binding.ChanReceiver(messageChannel), nil)
	ev := test.MinEvent()

	c, err := client.New(transport, client.WithoutTracePropagation())
	require.NoError(t, err)

	err = c.Send(context.Background(), ev)
	require.NoError(t, err)

	result := <-messageChannel

	test.AssertEventEquals(t, ev, event.Event(*(result.(*binding.EventMessage))))
}

func TestTransportReceive(t *testing.T) {
	messageChannel := make(chan binding.Message, 1)
	eventReceivedChannel := make(chan event.Event, 1)
	transport := bindings2.NewSendingTransport(binding.ChanSender(messageChannel), binding.ChanReceiver(messageChannel), nil)
	ev := test.MinEvent()

	c, err := client.New(transport)
	require.NoError(t, err)

	messageChannel <- (*binding.EventMessage)(&ev)

	go func() {
		err = c.StartReceiver(context.Background(), func(event event.Event) {
			eventReceivedChannel <- event
		})
		require.NoError(t, err)
	}()

	result := <-eventReceivedChannel

	test.AssertEventEquals(t, ev, result)
}
