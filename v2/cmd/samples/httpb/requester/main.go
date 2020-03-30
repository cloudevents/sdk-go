package main

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	for i := 0; i < 10; i++ {
		e := cloudevents.NewEvent()
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/v2/cmd/samples/httpb/requester")
		_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
			"id":      i,
			"message": "Hello, World!",
		})

		resp, result := c.Request(ctx, e)
		if resp != nil {
			//log.Printf("response: %s", resp)
		}

		if protocol.IsACK(result) {
			log.Printf("%d: ACK'ed  👍", i)
		} else if protocol.IsNACK(result) {
			log.Printf("%d: NACK'ed 😭 %s", i, result)
		} else {
			log.Printf("%d: Error %s", i, result)
		}
	}
}
