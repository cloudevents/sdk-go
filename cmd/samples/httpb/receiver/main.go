package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudevents/sdk-go/pkg/transport/http"

	cloudevents "github.com/cloudevents/sdk-go"
)

func main() {
	ctx := context.Background()
	p, err := cloudevents.NewHTTPProtocol()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	t, err := http.New(p)
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
