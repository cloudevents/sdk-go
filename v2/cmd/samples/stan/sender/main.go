package main

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	ce_stan "github.com/cloudevents/sdk-go/v2/protocol/stan"
	"log"
)

func main() {
	s, err := ce_stan.NewSender("test-cluster", "test-client", "test-subject", ce_stan.StanOptions())
	if err != nil {
		log.Fatalf("failed to create protocol: %v", err)
	}

	c, err := cloudevents.NewClient(s, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	for i := 0; i < 10; i++ {
		e := cloudevents.NewEvent()
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/v2/cmd/samples/stan/sender")
		_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
			"id":      i,
			"message": "Hello, World!",
		})

		err := c.Send(context.Background(), e)
		if err != nil {
			log.Printf("failed to send: %v", err)
		} else {
			log.Printf("sent: %d", i)
		}
	}
}
