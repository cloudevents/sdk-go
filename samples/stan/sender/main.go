/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"

	cestan "github.com/cloudevents/sdk-go/protocol/stan/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	s, err := cestan.NewSender("test-cluster", "test-client", "test-subject", cestan.StanOptions())
	if err != nil {
		log.Fatalf("failed to create protocol: %v", err)
	}

	defer s.Close(context.Background())

	c, err := cloudevents.NewClient(s, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	for i := 0; i < 10; i++ {
		e := cloudevents.NewEvent()
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/v2/samples/stan/sender")
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
