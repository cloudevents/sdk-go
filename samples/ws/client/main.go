/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"
	"sync"
	"sync/atomic"

	cews "github.com/cloudevents/sdk-go/protocol/ws/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	ctx := context.Background()
	p, err := cews.Dial(ctx, "http://localhost:8080", nil)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		for i := 0; i < 10; i++ {
			e := cloudevents.NewEvent()
			e.SetType("com.cloudevents.sample.sent")
			e.SetSource("https://github.com/cloudevents/sdk-go/v2/samples/stan/sender")
			_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
				"id":      i,
				"message": "Hello, World!",
			})

			if result := c.Send(context.Background(), e); cloudevents.IsUndelivered(result) {
				log.Printf("Failed to send: %v", err)
			} else {
				log.Printf("Sent: %d, accepted: %t", i, cloudevents.IsACK(result))
			}
		}
		wg.Done()
	}()
	go func() {
		received := uint32(0)
		ctx, cancel := context.WithCancel(ctx)
		err := c.StartReceiver(ctx, func(event cloudevents.Event) {
			log.Printf("Received event:\n%v", event)
			if atomic.AddUint32(&received, 1) == 10 {
				cancel()
			}
		})
		if err != nil {
			log.Printf("failed to start receiver: %v", err)
		} else {
			<-ctx.Done()
		}
		wg.Done()
	}()

	wg.Wait()
}
