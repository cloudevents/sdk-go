package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudevents/sdk-go/pkg/client"
	"github.com/cloudevents/sdk-go/pkg/event"
	cloudeventsnats "github.com/cloudevents/sdk-go/pkg/transport/nats"
	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	// NATSServer URL to connect to the nats server.
	NATSServer string `envconfig:"NATS_SERVER" default:"http://localhost:4222" required:"true"`

	// Subject is the nats subject to subscribe for cloudevents on.
	Subject string `envconfig:"SUBJECT" default:"sample" required:"true"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	ctx := context.Background()

	t, err := cloudeventsnats.New(env.NATSServer, env.Subject)
	if err != nil {
		log.Fatalf("failed to create nats transport, %s", err.Error())
	}
	c, err := client.New(t)
	if err != nil {
		log.Fatalf("failed to create client, %s", err.Error())
	}

	for {
		if err := c.StartReceiver(ctx, receive); err != nil {
			log.Printf("failed to start nats receiver, %s", err.Error())
		}
	}
}

type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func receive(ctx context.Context, event event.Event, resp *event.EventResponse) error {
	fmt.Printf("Got Event Context: %+v\n", event.Context)

	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}
	fmt.Printf("Got Data: %+v\n", data)

	fmt.Printf("----------------------------\n")
	return nil
}
