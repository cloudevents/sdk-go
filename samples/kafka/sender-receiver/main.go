/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

const (
	count = 10
)

func main() {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0

	// With NewProtocol you can use the same client both to send and receive.
	protocol, err := kafka_sarama.NewProtocol([]string{"127.0.0.1:9092"}, saramaConfig, "test-topic", "test-topic")
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	defer protocol.Close(context.Background())

	c, err := cloudevents.NewClient(protocol, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	// Create a done channel to block until we've received (count) messages
	done := make(chan struct{})

	// Start the receiver
	go func() {
		log.Printf("will listen consuming topic test-topic\n")
		var recvCount int32
		err = c.StartReceiver(context.TODO(), func(ctx context.Context, event cloudevents.Event) {
			receive(ctx, event)
			if atomic.AddInt32(&recvCount, 1) == count {
				done <- struct{}{}
			}
		})
		if err != nil {
			log.Fatalf("failed to start receiver: %s", err)
		} else {
			log.Printf("receiver stopped\n")
		}
	}()

	// Start sending the events
	for i := 0; i < count; i++ {
		e := cloudevents.NewEvent()
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/v2/samples/httpb/requester")
		_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
			"id":      i,
			"message": "Hello, World!",
		})

		if result := c.Send(context.Background(), e); cloudevents.IsUndelivered(result) {
			log.Printf("failed to send: %v", result)
		} else {
			log.Printf("sent: %d, accepted: %t", i, cloudevents.IsACK(result))
		}
	}

	<-done
}

func receive(ctx context.Context, event cloudevents.Event) {
	fmt.Printf("%s", event)
}
