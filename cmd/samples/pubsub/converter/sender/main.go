package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/kelseyhightower/envconfig"
)

/*

To view: gcloud pubsub subscriptions pull --auto-ack foo

To post: gcloud pubsub topics publish demo_cloudevents --message '{"id":123,"message":"hi from the terminal"}'

*/

type envConfig struct {
	ProjectID string `envconfig:"GOOGLE_CLOUD_PROJECT" required:"true"`

	TopicID string `envconfig:"PUBSUB_TOPIC" default:"demo_cloudevents" required:"true"`
}

// Basic data struct.
type Example struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, env.ProjectID)
	if err != nil {
		log.Fatal(err)
	}

	topic := client.Topic(env.TopicID)

	// Create the topic if it doesn't exist.
	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		log.Printf("Topic %v doesn't exist - creating it", env.TopicID)
		_, err = client.CreateTopic(ctx, env.TopicID)
		if err != nil {
			log.Fatal(err)
		}
	}

	data := &Example{
		ID:      123,
		Message: "Hello, World!",
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}

	msg := &pubsub.Message{
		Data: b,
	}

	if _, err := topic.Publish(ctx, msg).Get(ctx); err != nil {
		log.Printf("Could not publish message: %v", err)
	}
}
