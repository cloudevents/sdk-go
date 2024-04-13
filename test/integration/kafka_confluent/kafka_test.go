/*
 Copyright 2023 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_confluent

import (
	"context"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	confluent "github.com/cloudevents/sdk-go/protocol/kafka_confluent/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
)

const (
	TEST_GROUP_ID    = "test_confluent_group_id"
	BOOTSTRAP_SERVER = "localhost:9192"
)

type receiveEvent struct {
	event cloudevents.Event
	err   error
}

func TestSendEvent(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, eventIn event.Event) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		topicName := "test-ce-confluent-" + uuid.New().String()
		// create the topic with kafka.AdminClient manually
		admin, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": BOOTSTRAP_SERVER})
		require.NoError(t, err)

		_, err = admin.CreateTopics(ctx, []kafka.TopicSpecification{{
			Topic:             topicName,
			NumPartitions:     1,
			ReplicationFactor: 1}})
		require.NoError(t, err)

		eventIn = test.ConvertEventExtensionsToString(t, eventIn)

		// start a cloudevents receiver client go to receive the event
		eventChan := make(chan receiveEvent)

		receiverReady := make(chan bool)
		go func() {
			p, err := protocolFactory("", []string{topicName})
			if err != nil {
				eventChan <- receiveEvent{err: err}
				return
			}
			defer p.Close(ctx)

			client, err := cloudevents.NewClient(p)
			if err != nil {
				eventChan <- receiveEvent{err: err}
			}

			receiverReady <- true
			err = client.StartReceiver(ctx, func(event cloudevents.Event) {
				eventChan <- receiveEvent{event: event}
			})
			if err != nil {
				eventChan <- receiveEvent{err: err}
			}
		}()

		<-receiverReady

		// start a cloudevents sender client go to send the event
		p, err := protocolFactory(topicName, nil)
		require.NoError(t, err)
		defer p.Close(ctx)

		client, err := cloudevents.NewClient(p)
		require.NoError(t, err)
		res := client.Send(ctx, eventIn)
		require.NoError(t, res)

		// check the received event
		receivedEvent := <-eventChan
		require.NoError(t, receivedEvent.err)
		eventOut := test.ConvertEventExtensionsToString(t, receivedEvent.event)

		// test.AssertEventEquals(t, eventIn, receivedEvent.event)
		err = test.AllOf(
			test.HasExactlyAttributesEqualTo(eventIn.Context),
			test.HasData(eventIn.Data()),
			test.HasExtensionKeys([]string{confluent.KafkaPartitionKey, confluent.KafkaOffsetKey}),
			test.HasExtension(confluent.KafkaTopicKey, topicName),
		)(eventOut)
		require.NoError(t, err)
	})
}

// To start a local environment for testing:
// Option 1: Start it on port 9092
//
//	docker run --rm --net=host -p 9092:9092 confluentinc/confluent-local
//
// Option 2: Start it on port 9192
// docker run --rm \
// --name broker \
// --hostname broker \
// -p 9192:9192 \
// -e KAFKA_ADVERTISED_LISTENERS='PLAINTEXT://broker:29192,PLAINTEXT_HOST://localhost:9192' \
// -e KAFKA_CONTROLLER_QUORUM_VOTERS='1@broker:29193' \
// -e KAFKA_LISTENERS='PLAINTEXT://broker:29192,CONTROLLER://broker:29193,PLAINTEXT_HOST://0.0.0.0:9192' \
// confluentinc/confluent-local:latest
func protocolFactory(sendTopic string, receiveTopic []string,
) (*confluent.Protocol, error) {

	var p *confluent.Protocol
	var err error
	if receiveTopic != nil {
		p, err = confluent.New(confluent.WithConfigMap(&kafka.ConfigMap{
			"bootstrap.servers":  BOOTSTRAP_SERVER,
			"group.id":           TEST_GROUP_ID,
			"auto.offset.reset":  "earliest",
			"enable.auto.commit": "true",
		}), confluent.WithReceiverTopics(receiveTopic))
	}
	if sendTopic != "" {
		p, err = confluent.New(confluent.WithConfigMap(&kafka.ConfigMap{
			"bootstrap.servers":   BOOTSTRAP_SERVER,
			"go.delivery.reports": false,
		}), confluent.WithSenderTopic(sendTopic))
	}
	return p, err
}
