/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

// Package test provides re-usable functions for binding tests.
package test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

// SendReceive does s.Send(in), then it receives the message in r.Receive() and executes outAssert
// Halt test on error.
func SendReceive(t *testing.T, ctx context.Context, in binding.Message, s protocol.Sender, r protocol.Receiver, outAssert func(binding.Message)) {
	t.Helper()
	wg := sync.WaitGroup{}
	wg.Add(2)

	// 'wait' is used to ensure that we at least wait until the Receiver
	// thread starts. We'll then use a 'sleep' (hoping) that waits until
	// the Receiver itself is ready
	wait := make(chan bool)

	go func() {
		defer wg.Done()
		wait <- true
		out, result := r.Receive(ctx)
		if !protocol.IsACK(result) {
			require.NoError(t, result)
		}
		outAssert(out)
		finishErr := out.Finish(nil)
		require.NoError(t, finishErr)
	}()

	// Wait until receiver thread starts, and then wait a second to
	// give the "Receive" call a chance to start (finger's crossed)
	<-wait
	time.Sleep(time.Second)

	go func() {
		defer wg.Done()
		mx := sync.Mutex{}
		finished := false
		in = binding.WithFinish(in, func(err error) {
			require.NoError(t, err)
			mx.Lock()
			finished = true
			mx.Unlock()
		})
		result := s.Send(ctx, in)

		mx.Lock()
		if !protocol.IsACK(result) {
			require.NoError(t, result)
		}
		mx.Unlock()

		time.Sleep(5 * time.Millisecond) // let the receiver receive.
		mx.Lock()
		require.True(t, finished)
		mx.Unlock()
	}()

	wg.Wait()
}
