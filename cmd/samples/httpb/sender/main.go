package main

import (
	"context"
	"fmt"
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
		e.SetSource("https://github.com/cloudevents/sdk-go/cmd/samples/httpb/sender")
		_ = e.SetData(map[string]string{
			"sequence": fmt.Sprintf("%d", i),
			"message":  "Hello, World!",
		})

		if _, resp, err := c.Send(ctx, e); err != nil {
			log.Printf("failed to send: %v", err)
		} else if resp != nil {
			fmt.Printf("got back a response: \n%s", resp)
		} else {
			log.Printf("send: %d", i)
		}
	}
}
