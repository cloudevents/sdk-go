package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/pubsub"
	pscontext "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/pubsub/context"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

/*

curl -X POST -H "Content-Type: application/json"  -d '{"id":123,"message":"hello world"}' http://localhost:8080

*/

type envConfig struct {
	ProjectID      string `envconfig:"GOOGLE_CLOUD_PROJECT"`
	TopicID        string `envconfig:"PUBSUB_TOPIC" default:"demo_cloudevents" required:"true"`
	SubscriptionID string `envconfig:"PUBSUB_SUBSCRIPTION" default:"foo" required:"true"`
}

// Basic data struct.
type Example struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

func gotEvent(ctx context.Context, event cloudevents.Event) error {
	fmt.Printf("CloudEvent.Event: %+v\n", event)
	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Data Error: %s\n", err.Error())
	}
	fmt.Printf("Structured Data: %+v\n", data)

	fmt.Printf("Transport Context: %+v\n", cloudevents.HTTPTransportContextFrom(ctx))

	fmt.Printf("----------------------------\n")
	return nil
}

func convert(ctx context.Context, m transport.Message, err error) (*cloudevents.Event, error) {
	log.Printf("trying to recover from %v", err)

	if msg, ok := m.(*pubsub.Message); ok {
		tx := pscontext.TransportContextFrom(ctx)
		// Make a new event and convert the message payload.
		event := cloudevents.NewEvent()
		event.SetSource("github.com/cloudevents/cmd/samples/pubsub/converter/receiver")
		event.SetType(fmt.Sprintf("io.cloudevents.converter.pubsub.%s", strings.ToLower(tx.Method)))
		event.SetID(uuid.New().String())
		event.Data = msg.Data

		return &event, nil
	}
	return nil, err
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	ctx := context.Background()

	t, err := pubsub.New(context.Background(),
		pubsub.WithProjectID(env.ProjectID),
		pubsub.WithTopicID(env.TopicID),
		pubsub.WithSubscriptionID(env.SubscriptionID))
	if err != nil {
		log.Printf("failed to create transport, %v", err)
		os.Exit(1)
	}
	c, err := cloudevents.NewClient(t, cloudevents.WithConverterFn(convert))
	if err != nil {
		log.Printf("failed to create client, %v", err)
		os.Exit(1)
	}

	log.Printf("will listen on %s/%s\n", env.TopicID, env.SubscriptionID)
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, gotEvent))
}
