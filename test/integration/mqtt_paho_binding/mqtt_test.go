/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package mqtt_paho_binding

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/protocol/mqtt_paho/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/assert"

	. "github.com/cloudevents/sdk-go/v2/binding/test"
	. "github.com/cloudevents/sdk-go/v2/test"
)

type receiveMessage struct {
	msg binding.Message
	err error
}

func TestEncodingBinding(t *testing.T) {
	testCases := []struct {
		name             string
		inEncoding       binding.Encoding
		inEventEncoder   func(tt *testing.T, inEvent event.Event) binding.Message
		expectedEncoding binding.Encoding
	}{
		{
			name:       "Structured message with Structured encoding",
			inEncoding: binding.EncodingStructured,
			inEventEncoder: func(tt *testing.T, inEvent event.Event) binding.Message {
				return MustCreateMockStructuredMessage(tt, inEvent)
			},
			expectedEncoding: binding.EncodingStructured,
		},
		{
			name:       "Binary message with Binary encoding",
			inEncoding: binding.EncodingBinary,
			inEventEncoder: func(tt *testing.T, inEvent event.Event) binding.Message {
				return MustCreateMockBinaryMessage(inEvent)
			},
			expectedEncoding: binding.EncodingBinary,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			EachEvent(t, Events(), func(t *testing.T, inEvent event.Event) {
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				topicName := "test-ce-client-" + uuid.New().String()

				// start a cloudevents receiver client go to receive the event
				receiveChan := make(chan receiveMessage)
				defer close(receiveChan)
				startReceiver(ctx, topicName, receiveChan)

				// start a cloudevents sender client go to send the event
				sender, err := getProtocol(ctx, topicName)
				require.NoError(t, err)
				defer sender.Close(ctx)

				inEvent = ConvertEventExtensionsToString(t, inEvent)
				inMessage := tt.inEventEncoder(t, inEvent)
				ctx = binding.WithPreferredEventEncoding(ctx, tt.inEncoding)

				timer := time.NewTimer(5 * time.Millisecond)
				defer timer.Stop()
				for {
					select {
					case <-ctx.Done():
						require.Fail(t, "timeout waiting for event")
						return
					case receiveMsg := <-receiveChan:
						require.NoError(t, receiveMsg.err)
						outMessage := receiveMsg.msg
						assert.Equal(t, tt.expectedEncoding, outMessage.ReadEncoding())
						outEvent := MustToEvent(t, ctx, outMessage)
						AssertEventEquals(t, inEvent, ConvertEventExtensionsToString(t, outEvent))
						return
					case <-timer.C:
						result := sender.Send(ctx, inMessage)
						require.NoError(t, result)
						// the receiver mightn't be ready before the sender send the message, so we retry
						continue
					}
				}
			})
		})
	}
}

func startReceiver(ctx context.Context, topicName string, messageChan chan receiveMessage) {
	receiver, err := getProtocol(ctx, topicName)
	if err != nil {
		messageChan <- receiveMessage{err: err}
	}

	// Used to try to make sure the receiver is ready before we start to
	// get events
	wait := make(chan bool)

	go func() {
		wait <- true
		err := receiver.OpenInbound(ctx)
		if err != nil {
			messageChan <- receiveMessage{err: err}
		}
		receiver.Close(ctx)
	}()

	// Wait for other thread to start and run OpenInbound + sleep a sec
	// hoping that things will get ready before we call Receive() below
	<-wait
	time.Sleep(time.Second)

	go func() {
		msg, result := receiver.Receive(ctx)
		messageChan <- receiveMessage{msg, result}
	}()
}

func getProtocol(ctx context.Context, topic string) (*mqtt_paho.Protocol, error) {
	broker := "127.0.0.1:1883"

	conn, err := net.Dial("tcp", broker)
	if err != nil {
		return nil, err
	}
	config := &paho.ClientConfig{
		Conn: conn,
	}
	publishOpt := &paho.Publish{
		Topic: topic, QoS: 0,
	}
	subscribeOpt := &paho.Subscribe{
		Subscriptions: map[string]paho.SubscribeOptions{
			topic: {QoS: 0},
		},
	}

	p, err := mqtt_paho.New(ctx, config, mqtt_paho.WithPublish(publishOpt), mqtt_paho.WithSubscribe(subscribeOpt))
	return p, err
}
