package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudevents/sdk-go"
)

func main() {
	ctx := context.Background()

	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("will listen on :8080\n")
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, gotEvent))
}

// Example is the expected incoming event.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func gotEvent(ctx context.Context, event cloudevents.Event) {
	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("failed to get data as Example: %s\n", err.Error())
		return
	}

	fmt.Printf("%s", event)
	fmt.Printf("%s\n", cloudevents.HTTPTransportContextFrom(ctx))
}
