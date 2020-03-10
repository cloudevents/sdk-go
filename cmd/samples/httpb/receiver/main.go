package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudevents/sdk-go/pkg/transport/httpb"

	cloudevents "github.com/cloudevents/sdk-go"
)

func main() {
	ctx := context.Background()

	t, err := httpb.New()
	if err != nil {
		log.Fatalf("failed to create transport, %v", err)
	}

	c, err := cloudevents.NewClient(t)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("will listen on :8080\n")
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, receive))
}

func receive(ctx context.Context, event cloudevents.Event) {
	fmt.Printf("%s", event)
}
