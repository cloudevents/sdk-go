package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventspubsub "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/pubsub"
	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	ProjectID string `envconfig:"GOOGLE_CLOUD_PROJECT"`

	TopicID string `envconfig:"PUBSUB_TOPIC" default:"demo_cloudevents" required:"true"`

	SubscriptionID string `envconfig:"PUBSUB_SUBSCRIPTION" default:"foo" requried:"true"`
}

type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func receive(ctx context.Context, event cloudevents.Event, resp *cloudevents.EventResponse) error {
	fmt.Printf("Event Context: %+v\n", event.Context)

	fmt.Printf("Transport Context: %+v\n", cloudeventspubsub.TransportContextFrom(ctx))

	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}
	fmt.Printf("Data: %+v\n", data)

	fmt.Printf("----------------------------\n")
	return nil
}

func main() {
	ctx := context.Background()

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}

	t, err := cloudeventspubsub.New(context.Background(),
		cloudeventspubsub.WithProjectID(env.ProjectID),
		cloudeventspubsub.WithTopicID(env.TopicID),
		cloudeventspubsub.WithSubscriptionID(env.SubscriptionID))
	if err != nil {
		log.Fatalf("failed to create pubsub transport, %s", err.Error())
	}
	c, err := client.New(t)
	if err != nil {
		log.Fatalf("failed to create client, %s", err.Error())
	}

	log.Println("Created client, listening...")

	if err := c.StartReceiver(ctx, receive); err != nil {
		log.Fatalf("failed to start nats receiver, %s", err.Error())
	}
}
