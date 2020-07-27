package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/test"
)

func TestMockSenderClient(t *testing.T) {
	client, eventCh := NewMockSenderClient(t, 1)
	event := test.FullEvent()

	require.NoError(t, client.Send(context.TODO(), event))
	received := <-eventCh
	test.AssertEventEquals(t, event, received)
}

func TestMockRequesterClient(t *testing.T) {
	replyEvent := test.MinEvent()

	client, eventCh := NewMockRequesterClient(t, 1, func(inMessage event.Event) (*event.Event, protocol.Result) {
		return &replyEvent, nil
	})
	inputEvent := test.FullEvent()

	require.NoError(t, client.Send(context.TODO(), inputEvent))
	received := <-eventCh
	test.AssertEventEquals(t, inputEvent, received)

	actualReply, err := client.Request(context.TODO(), inputEvent)
	require.NoError(t, err)
	received = <-eventCh
	test.AssertEventEquals(t, inputEvent, received)
	if actualReply != nil {
		test.AssertEventEquals(t, replyEvent, *actualReply)
	} else {
		t.Fatalf("Expected reply, but got nil")
	}
}

func TestMockReceiverClient(t *testing.T) {
	client, eventCh := NewMockReceiverClient(t, 1)
	inputEvent := test.FullEvent()

	eventsReceivedInTheReceiver := make(chan event.Event, 1)

	go func() {
		require.NoError(t, client.StartReceiver(context.TODO(), func(e event.Event) {
			eventsReceivedInTheReceiver <- e
		}))
	}()

	eventCh <- inputEvent
	received := <-eventsReceivedInTheReceiver
	test.AssertEventEquals(t, inputEvent, received)
}

func TestMockResponderClient(t *testing.T) {
	client, inEventCh, outEventCh := NewMockResponderClient(t, 1)
	inputEvent := test.FullEvent()
	replyEvent := test.MinEvent()

	eventsReceivedInTheReceiver := make(chan event.Event, 1)

	go func() {
		require.NoError(t, client.StartReceiver(context.TODO(), func(e event.Event) (*event.Event, protocol.Result) {
			eventsReceivedInTheReceiver <- e
			return &replyEvent, protocol.NewResult("OK")
		}))
	}()

	inEventCh <- inputEvent
	received := <-eventsReceivedInTheReceiver
	test.AssertEventEquals(t, inputEvent, received)

	reply := <-outEventCh
	require.Equal(t, protocol.NewResult("OK"), reply.Result)
	test.AssertEventEquals(t, replyEvent, reply.Event)
}
