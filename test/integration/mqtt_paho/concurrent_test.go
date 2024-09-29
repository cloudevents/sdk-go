/*
Copyright 2024 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package mqtt_paho

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
)

func TestConcurrentSendingEvent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	topicName := "test-ce-client-" + uuid.New().String()

	readyCh := make(chan bool)
	defer close(readyCh)

	senderNum := 10  // 10 gorutine to sending the events
	eventNum := 1000 // each gorutine sender publishs 1,000 events

	var g errgroup.Group

	// start a receiver
	c, err := cloudevents.NewClient(protocolFactory(ctx, t, topicName), cloudevents.WithUUIDs())
	require.NoError(t, err)
	g.Go(func() error {
		// verify all of events can be recieved
		count := senderNum * eventNum
		var mu sync.Mutex
		return c.StartReceiver(ctx, func(event cloudevents.Event) {
			mu.Lock()
			defer mu.Unlock()
			count--
			if count == 0 {
				readyCh <- true
			}
		})
	})
	// wait for 5 seconds to ensure the receiver starts safely
	time.Sleep(5 * time.Second)

	// start a sender client to pulish events concurrently
	client, err := cloudevents.NewClient(protocolFactory(ctx, t, topicName), cloudevents.WithUUIDs())
	require.NoError(t, err)

	evt := cloudevents.NewEvent()
	evt.SetType("com.cloudevents.sample.sent")
	evt.SetSource("concurrent-sender")
	err = evt.SetData(cloudevents.ApplicationJSON, map[string]interface{}{"message": "Hello, World!"})
	require.NoError(t, err)

	for i := 0; i < senderNum; i++ {
		g.Go(func() error {
			for j := 0; j < eventNum; j++ {
				result := client.Send(
					cecontext.WithTopic(ctx, topicName),
					evt,
				)
				if result != nil {
					return result
				}
			}
			return nil
		})
	}

	// wait until all the events are received
	handleEvent(ctx, readyCh, cancel, t)

	require.NoError(t, g.Wait())
}

func handleEvent(ctx context.Context, readyCh <-chan bool, cancel context.CancelFunc, t *testing.T) {
	for {
		select {
		case <-ctx.Done():
			require.Fail(t, "Test failed: timeout reached before events were received")
			return
		case <-readyCh:
			cancel()
			t.Logf("Test passed: events successfully received")
			return
		}
	}
}
