/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	cepubsub "github.com/cloudevents/sdk-go/protocol/pubsub/v2"
	pscontext "github.com/cloudevents/sdk-go/protocol/pubsub/v2/context"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/kelseyhightower/envconfig"
)

/*
To setup:
gcloud pubsub topics create demo_cloudevents
gcloud pubsub subscriptions create foo --topic=demo_cloudevents
To test:
gcloud pubsub topics publish demo_cloudevents --message='{"Hello": "world"}'
To fix a bad message:
gcloud pubsub subscriptions pull --auto-ack foo
*/

type envConfig struct {
	ProjectID string `envconfig:"GOOGLE_CLOUD_PROJECT" required:"true"`

	TopicID string `envconfig:"PUBSUB_TOPIC" default:"demo_cloudevents" required:"true"`

	SubscriptionID string `envconfig:"PUBSUB_SUBSCRIPTION" default:"foo" required:"true"`
}

type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func receive(ctx context.Context, event event.Event) error {
	fmt.Printf("Event Context: %+v\n", event.Context)

	fmt.Printf("Protocol Context: %+v\n", pscontext.ProtocolContextFrom(ctx))

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

	t, err := cepubsub.New(context.Background(),
		cepubsub.WithProjectID(env.ProjectID),
		cepubsub.WithTopicID(env.TopicID),
		cepubsub.WithSubscriptionID(env.SubscriptionID))
	if err != nil {
		log.Fatalf("failed to create pubsub protocol, %s", err.Error())
	}
	c, err := cloudevents.NewClient(t)

	if err != nil {
		log.Fatalf("failed to create client, %s", err.Error())
	}

	log.Println("Created client, listening...")

	if err := c.StartReceiver(ctx, receive); err != nil {
		log.Fatalf("failed to start pubsub receiver, %s", err.Error())
	}
}
