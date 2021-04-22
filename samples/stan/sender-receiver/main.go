/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"

	cestan "github.com/cloudevents/sdk-go/protocol/stan/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	protocol, err := cestan.NewProtocol("test-cluster", "test-client", "send-test-subject", "receiver-test-subject",
		cestan.StanOptions())
	if err != nil {
		log.Fatalf("failed to create protocol: %v", err)
	}

	defer protocol.Close(context.Background())

	c, err := cloudevents.NewClient(protocol, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// Start the receiver
	go func() {
		log.Printf("will listen consuming topic test-topic\n")
		err = c.StartReceiver(context.TODO(), receive)
		if err != nil {
			log.Fatalf("failed to start receiver: %s", err)
		} else {
			log.Printf("receiver stopped\n")
		}
	}()

	// Start sending the events
	for i := 0; i < 10; i++ {
		e := cloudevents.NewEvent()
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/v2/samples/stan/sender_and_receiver")
		_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
			"id":      i,
			"message": "Hello, World!",
		})

		if result := c.Send(context.Background(), e); cloudevents.IsUndelivered(result) {
			log.Printf("failed to send: %v", err)
		} else {
			log.Printf("sent: %d, accepted: %t", i, cloudevents.IsACK(result))
		}
	}
}

func receive(ctx context.Context, event cloudevents.Event) {
	fmt.Printf("%s", event)
}
