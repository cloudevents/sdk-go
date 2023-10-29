/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package mqtt_paho

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	mqtt_paho "github.com/cloudevents/sdk-go/protocol/mqtt_paho/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
)

type receiveEvent struct {
	event cloudevents.Event
	err   error
}

func TestSendEvent(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, inEvent event.Event) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		topicName := "test-ce-client-" + uuid.New().String()
		inEvent = test.ConvertEventExtensionsToString(t, inEvent)

		// start a cloudevents receiver client go to receive the event
		eventChan := make(chan receiveEvent)
		defer close(eventChan)

		// 'wait' is used to ensure that we at least wait until the Receiver
		// thread starts. We'll then use a 'sleep' (hoping) that waits until
		// the Receiver itself is ready
		wait := make(chan bool)

		go func() {
			client, err := cloudevents.NewClient(protocolFactory(ctx, t, topicName))
			if err != nil {
				eventChan <- receiveEvent{err: err}
				return
			}
			wait <- true
			err = client.StartReceiver(ctx, func(event cloudevents.Event) {
				eventChan <- receiveEvent{event: event}
			})
			if err != nil {
				eventChan <- receiveEvent{err: err}
				return
			}
		}()

		// Wait until receiver thread starts, and then wait a second to
		// give the "StartReceive" call a chance to start (finger's crossed)
		<-wait
		time.Sleep(time.Second)

		// start a cloudevents sender client go to send the event, set the topic on context
		client, err := cloudevents.NewClient(protocolFactory(ctx, t, ""))
		require.NoError(t, err)

		timer := time.NewTimer(5 * time.Millisecond)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				require.Fail(t, "timeout waiting for event")
				return
			case eventOut := <-eventChan:
				require.NoError(t, eventOut.err)
				test.AssertEventEquals(t, inEvent, test.ConvertEventExtensionsToString(t, eventOut.event))
				return
			case <-timer.C:
				result := client.Send(
					cecontext.WithTopic(ctx, topicName),
					inEvent,
				)
				require.NoError(t, result)
				// the receiver mightn't be ready before the sender send the message, so wait and we retry
				continue
			}
		}
	})
}

// To start a local environment for testing:
// docker run -it --rm --name mosquitto -p 1883:1883 eclipse-mosquitto:2.0 mosquitto -c /mosquitto-no-auth.conf
// the protocolFactory will generate a unique connection clientId when it be invoked
func protocolFactory(ctx context.Context, t testing.TB, topicName string) *mqtt_paho.Protocol {
	broker := "127.0.0.1:1883"
	conn, err := net.Dial("tcp", broker)
	require.NoError(t, err)
	config := &paho.ClientConfig{
		Conn: conn,
	}
	connOpt := &paho.Connect{
		KeepAlive:  30,
		CleanStart: true,
	}
	publishOpt := &paho.Publish{
		Topic: topicName, QoS: 0,
	}
	subscribeOpt := &paho.Subscribe{
		Subscriptions: map[string]paho.SubscribeOptions{
			topicName: {QoS: 0},
		},
	}

	p, err := mqtt_paho.New(ctx, config, mqtt_paho.WithConnect(connOpt), mqtt_paho.WithPublish(publishOpt), mqtt_paho.WithSubscribe(subscribeOpt))
	require.NoError(t, err)

	return p
}
