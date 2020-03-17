package main

import (
	"context"
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// Basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func main() {
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	for i := 0; i < 10; i++ {
		data := &Example{
			Sequence: i,
			Message:  "Hello, World!",
		}

		var version string
		switch i % 2 {
		case 0:
			version = cloudevents.VersionV03
		case 1:
			version = cloudevents.VersionV1
		}

		event := cloudevents.NewEvent(version)
		event.SetType("com.cloudevents.sample.sent")
		event.SetSource("https://github.com/cloudevents/sdk-go/v2/cmd/samples/sender")
		_ = event.SetData(cloudevents.ApplicationJSON, data)

		if resp, err := c.Request(ctx, event); err != nil {
			log.Printf("failed to send: %v", err)
		} else if resp != nil {
			fmt.Printf("got back a response: \n%s", resp)
		} else {
			log.Printf("%s: %d - %s", event.Context.GetType(), data.Sequence, data.Message)
		}
	}
}
