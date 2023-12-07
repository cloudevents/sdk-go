/*
 Copyright 2023 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"

	confluent "github.com/cloudevents/sdk-go/protocol/kafka_confluent/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const topic = "test-confluent-topic"

func main() {
	ctx := context.Background()

	receiver, err := confluent.New(confluent.WithConfigMap(&kafka.ConfigMap{
		"bootstrap.servers":  "127.0.0.1:9092",
		"group.id":           "test-confluent-group-id",
		"auto.offset.reset":  "earliest", // only validated when the consumer group offset has saved before
		"enable.auto.commit": "true",
	}), confluent.WithReceiverTopics([]string{topic}))

	if err != nil {
		log.Fatalf("failed to create receiver, %v", err)
	}
	defer receiver.Close(ctx)

	// Setting the 'client.WithPollGoroutines(1)' to make sure the events from kafka partition are processed in order
	c, err := cloudevents.NewClient(receiver, client.WithPollGoroutines(1))
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("will listen consuming topic %s\n", topic)
	err = c.StartReceiver(ctx, receive)
	if err != nil {
		log.Fatalf("failed to start receiver: %s", err)
	} else {
		log.Printf("receiver stopped\n")
	}
}

func receive(ctx context.Context, event cloudevents.Event) {
	fmt.Printf("%s", event)
}
