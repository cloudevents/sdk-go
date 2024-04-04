/*
 Copyright 2023 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"

	confluent "github.com/cloudevents/sdk-go/protocol/kafka_confluent/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

const (
	count = 10
	topic = "test-confluent-topic"
)

func main() {
	ctx := context.Background()

	sender, err := confluent.New(confluent.WithConfigMap(&kafka.ConfigMap{
		"bootstrap.servers": "127.0.0.1:9092",
	}), confluent.WithSenderTopic(topic))

	defer sender.Close(ctx)

	c, err := cloudevents.NewClient(sender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	for i := 0; i < count; i++ {
		e := cloudevents.NewEvent()
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/samples/kafka_confluent/sender")
		_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
			"id":      i,
			"message": "Hello, World!",
		})

		if result := c.Send(
			// Set the producer message key
			confluent.WithMessageKey(ctx, e.ID()),
			e,
		); cloudevents.IsUndelivered(result) {
			log.Printf("failed to send: %v", result)
		} else {
			log.Printf("sent: %d, accepted: %t", i, cloudevents.IsACK(result))
		}
	}
}
