package main

import (
	"context"
	"log"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/transport/http"
)

func main() {
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	p, err := cloudevents.NewHTTPProtocol()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	t, err := http.New(p)
	if err != nil {
		log.Fatalf("failed to create transport, %v", err)
	}

	c, err := cloudevents.NewClient(t, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	for i := 0; i < 10; i++ {
		e := cloudevents.NewEvent()
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/cmd/samples/httpb/sender")
		_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
			"id":      i,
			"message": "Hello, World!",
		})

		err := c.Send(ctx, e)
		if err != nil {
			log.Printf("failed to send: %v", err)
		} else {
			log.Printf("sent: %d", i)
		}
	}
}
