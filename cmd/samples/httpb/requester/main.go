package main

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/event"
	"log"

	"github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/httpb"
)

func main() {
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")
	t, err := httpb.New()
	if err != nil {
		log.Fatalf("failed to create transport, %v", err)
	}

	c, err := cloudevents.NewClient(t, cloudevents.WithTimeNow(), cloudevents.WithUUIDs(), cloudevents.WithDataContentType(event.ApplicationJSON))
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	for i := 0; i < 10; i++ {
		e := cloudevents.NewEvent()
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/cmd/samples/httpb/requester")
		_ = e.SetData(map[string]interface{}{
			"id":      i,
			"message": "Hello, World!",
		})

		resp, err := c.Request(ctx, e)
		if err != nil {
			log.Printf("failed to send: %v", err)
		} else {
			log.Printf("sent: %d", i)
			if resp != nil {
				log.Printf("response: %s", resp)
			}
		}
	}
}
