package main

import (
	"context"
	"log"
	"os"

	"github.com/cloudevents/sdk-go"
	cepubsub "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/pubsub"
	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	ProjectID string `envconfig:"GOOGLE_CLOUD_PROJECT" required:"true"`

	TopicID string `envconfig:"PUBSUB_TOPIC" default:"demo_cloudevents" required:"true"`
}

// Basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}

	t, err := cepubsub.New(context.Background(),
		cepubsub.WithProjectID(env.ProjectID),
		cepubsub.WithTopicID(env.TopicID))
	if err != nil {
		log.Printf("failed to create pubsub transport, %s", err.Error())
		os.Exit(1)
	}
	c, err := cloudevents.NewClient(t, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Printf("failed to create client, %s", err.Error())
		os.Exit(1)
	}

	event := cloudevents.NewEvent(cloudevents.VersionV03)
	event.SetType("com.cloudevents.sample.sent")
	event.SetSource("github.com/cloudevents/sdk-go/cmd/samples/pubsub/sender/")
	_ = event.SetData(&Example{
		Sequence: 0,
		Message:  "HELLO",
	})

	_, _, err = c.Send(context.Background(), event)

	if err != nil {
		log.Printf("failed to send: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
