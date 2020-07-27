package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/gochan"
)

// NewMockSenderClient returns a client that can Send() event.
// All sent messages are delivered to the returned channel.
func NewMockSenderClient(t *testing.T, chanSize int, opts ...client.Option) (client.Client, <-chan event.Event) {
	require.NotZero(t, chanSize)

	eventCh := make(chan event.Event, chanSize)
	messageCh := make(chan binding.Message)

	// Output piping
	go func(messageCh <-chan binding.Message, eventCh chan<- event.Event) {
		for m := range messageCh {
			e, err := binding.ToEvent(context.TODO(), m)
			require.NoError(t, err)
			eventCh <- *e
		}
	}(messageCh, eventCh)

	c, err := client.New(gochan.Sender(messageCh), opts...)
	require.NoError(t, err)

	return c, eventCh
}

// NewMockRequesterClient returns a client that can perform Send() event and Request() event.
// All sent messages are delivered to the returned channel.
func NewMockRequesterClient(t *testing.T, chanSize int, replierFn func(inMessage event.Event) (*event.Event, protocol.Result), opts ...client.Option) (client.Client, <-chan event.Event) {
	require.NotZero(t, chanSize)
	require.NotNil(t, replierFn)

	eventCh := make(chan event.Event, chanSize)
	messageCh := make(chan binding.Message)

	replier := func(inMessage binding.Message) (binding.Message, error) {
		inEvent, err := binding.ToEvent(context.TODO(), inMessage)
		require.NoError(t, err)
		outEvent, err := replierFn(*inEvent)
		if outEvent != nil {
			return binding.ToMessage(outEvent), err
		}
		return nil, err
	}

	chanRequester := gochan.Requester{
		Ch:    messageCh,
		Reply: replier,
	}
	// Output piping
	go func(messageCh <-chan binding.Message, eventCh chan<- event.Event) {
		for m := range messageCh {
			e, err := binding.ToEvent(context.TODO(), m)
			require.NoError(t, err)
			eventCh <- *e
		}
	}(messageCh, eventCh)

	c, err := client.New(&chanRequester, opts...)
	require.NoError(t, err)

	return c, eventCh
}

// NewMockReceiverClient returns a client that can Receive events, without replying.
// The returned channel is the channel for sending messages to the client
func NewMockReceiverClient(t *testing.T, chanSize int, opts ...client.Option) (client.Client, chan<- event.Event) {
	require.NotZero(t, chanSize)

	eventCh := make(chan event.Event, chanSize)
	messageCh := make(chan binding.Message)

	// Input piping
	go func(messageCh chan<- binding.Message, eventCh <-chan event.Event) {
		for e := range eventCh {
			messageCh <- binding.ToMessage(&e)
		}
	}(messageCh, eventCh)

	c, err := client.New(gochan.Receiver(messageCh), opts...)
	require.NoError(t, err)

	return c, eventCh
}

type ClientMockResponse struct {
	Event  event.Event
	Result protocol.Result
}

// NewMockResponderClient returns a client that can Receive events and reply.
// The first returned channel is the channel for sending messages to the client, while the second one
// contains the eventual responses.
func NewMockResponderClient(t *testing.T, chanSize int, opts ...client.Option) (client.Client, chan<- event.Event, <-chan ClientMockResponse) {
	require.NotZero(t, chanSize)

	inEventCh := make(chan event.Event, chanSize)
	inMessageCh := make(chan binding.Message)

	outEventCh := make(chan ClientMockResponse, chanSize)
	outMessageCh := make(chan gochan.ChanResponderResponse)

	// Input piping
	go func(messageCh chan<- binding.Message, eventCh <-chan event.Event) {
		for e := range eventCh {
			messageCh <- binding.ToMessage(&e)
		}
	}(inMessageCh, inEventCh)

	// Output piping
	go func(messageCh <-chan gochan.ChanResponderResponse, eventCh chan<- ClientMockResponse) {
		for m := range messageCh {
			if m.Message != nil {
				e, err := binding.ToEvent(context.TODO(), m.Message)
				require.NoError(t, err)
				require.NoError(t, m.Message.Finish(nil))
				eventCh <- ClientMockResponse{
					Event:  *e,
					Result: m.Result,
				}
			}
			eventCh <- ClientMockResponse{Result: m.Result}
		}
	}(outMessageCh, outEventCh)

	c, err := client.New(&gochan.Responder{In: inMessageCh, Out: outMessageCh}, opts...)
	require.NoError(t, err)

	return c, inEventCh, outEventCh
}
