/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

// Package test provides Client test helpers.
package test

import (
	"context"
	"sync"
	"testing"

	"github.com/cloudevents/sdk-go/v2/protocol"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
)

// SendReceive does client.Send(in), then it receives the message using client.StartReceiver() and executes outAssert
// Halt test on error.
func SendReceive(t *testing.T, protocolFactory func() interface{}, in event.Event, outAssert func(e event.Event), opts ...client.Option) {
	t.Helper()
	pf := protocolFactory()
	c, err := client.New(pf, opts...)
	require.NoError(t, err)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		ctx, cancel := context.WithCancel(context.TODO())
		inCh := make(chan event.Event)
		defer func(channel chan event.Event) {
			cancel()
			close(channel)
			wg.Done()
		}(inCh)
		go func(channel chan event.Event) {
			err := c.StartReceiver(ctx, func(e event.Event) {
				channel <- e
			})
			if err != nil {
				require.NoError(t, err)
			}
		}(inCh)
		e := <-inCh
		outAssert(e)
	}()

	go func() {
		defer wg.Done()
		err := c.Send(context.Background(), in)
		require.NoError(t, err)
	}()

	wg.Wait()

	if closer, ok := pf.(protocol.Closer); ok {
		require.NoError(t, closer.Close(context.TODO()))
	}
}
