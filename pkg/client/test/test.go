package test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/client"
	"github.com/cloudevents/sdk-go/pkg/event"
)

// SendReceive does client.Send(in), then it receives the message using client.StartReceiver() and executes outAssert
// Halt test on error.
func SendReceive(t *testing.T, protocol interface{}, in event.Event, outAssert func(e event.Event), opts ...client.Option) {
	t.Helper()
	c, err := client.New(protocol, opts...)
	require.NoError(t, err)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		ctx, cancel := context.WithCancel(context.TODO())
		ch := make(chan event.Event)
		defer func() {
			cancel()
			close(ch)
			wg.Done()
		}()
		go func(channel chan event.Event) {
			err := c.StartReceiver(ctx, func(e event.Event) {
				channel <- e
			})
			if err != nil {
				require.NoError(t, err)
			}
		}(ch)
		e := <-ch
		outAssert(e)
	}()

	go func() {
		defer wg.Done()
		err := c.Send(context.Background(), in)
		require.NoError(t, err)
	}()

	wg.Wait()
}
