package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudevents/sdk-go"
)

var source = cloudevents.ParseURLRef("https://github.com/cloudevents/sdk-go/cmd/samples/sender")

// Basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

var subject = "this_thing"

func main() {
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	ctx = cloudevents.ContextWithHeader(ctx, "demo", "header value")

	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	for i := 0; i < 10; i++ {
		data := &Example{
			Sequence: i,
			Message:  "Hello, World!",
		}
		event := cloudevents.Event{
			Context: cloudevents.EventContextV03{
				Type:    "com.cloudevents.sample.sent",
				Source:  *source,
				Subject: &subject,
			}.AsV03(),
			Data: data,
		}

		if resp, err := c.Send(ctx, event); err != nil {
			log.Printf("failed to send: %v", err)
		} else if resp != nil {
			fmt.Printf("got back a response: \n%s", resp)
		} else {
			log.Printf("%s: %d - %s", event.Context.GetType(), data.Sequence, data.Message)
		}
	}
}
