/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"
	"sync"

	confluent "github.com/cloudevents/sdk-go/protocol/kafka_confluent/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

const (
	count = 1
	topic = "test-confluent-topic"
)

func main() {
	ctx := context.Background()

	sender, err := confluent.New(confluent.WithConfigMap(&kafka.ConfigMap{
		"bootstrap.servers": "127.0.0.1:9092",
	}), confluent.WithSenderTopic(topic))
	if err != nil {
		log.Fatalf("failed to create protocol, %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Listen to all the events on the default events channel
	// It's important to read these events otherwise the events channel will eventually fill up
	go func() {
		defer wg.Done()
		eventChan, err := sender.Events()
		if err != nil {
			log.Fatalf("failed to get events channel for sender, %v", err)
		}
		for e := range eventChan {
			switch ev := e.(type) {
			case *kafka.Message:
				// The message delivery report, indicating success or
				// permanent failure after retries have been exhausted.
				// Application level retries won't help since the client
				// is already configured to do that.
				m := ev
				if m.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
				} else {
					log.Printf("Delivered message to topic %s [%d] at offset %v\n",
						*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
				}
			case kafka.Error:
				// Generic client instance-level errors, such as
				// broker connection failures, authentication issues, etc.
				//
				// These errors should generally be considered informational
				// as the underlying client will automatically try to
				// recover from any errors encountered, the application
				// does not need to take action on them.
				log.Printf("Error: %v\n", ev)
			default:
				log.Printf("Ignored event: %v\n", ev)
			}
		}
	}()

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
		ctx = cloudevents.WithEncodingStructured(ctx)
		if result := c.Send(confluent.WithMessageKey(ctx, e.ID()), e); cloudevents.IsUndelivered(result) {
			log.Printf("failed to send: %v", result)
		} else {
			log.Printf("sent: %d, accepted: %t", i, cloudevents.IsACK(result))
		}
	}

	sender.Close(ctx)
	wg.Wait()
}
