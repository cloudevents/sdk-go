package main

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/v2/client"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	ctx := context.Background()
	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClientObserved(p, cloudevents.WithTracePropagation)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("will listen on :8080\n")
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, receive))
}

func receive(ctx context.Context, e cloudevents.Event) {
	ctx, span := client.TraceSpan(ctx, e)
	defer span.End()

	fmt.Printf("%s", e)
}
