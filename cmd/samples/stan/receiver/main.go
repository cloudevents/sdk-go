package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventsnats "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/stan"
	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	// ClusterID is which cluster to join
	ClusterID string `envconfig:"STAN_CLUSTER_ID" default:"test-cluster" required:"true"`
	// ClientID is which client id to identify yourself as
	ClientID string `envconfig:"STAN_CLIENT_ID" default:"stan-receiver" required:"true"`
	// Subject is the nats subject to subscribe for cloudevents on.
	Subject string `envconfig:"SUBJECT" default:"subject" required:"true"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	os.Exit(_main(os.Args[1:], env))
}

// Example represents a data payload
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func receive(ctx context.Context, event cloudevents.Event, resp *cloudevents.EventResponse) error {
	fmt.Printf("Got Event Context: %+v\n", event.Context)

	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}
	fmt.Printf("Got Data: %+v\n", data)

	fmt.Printf("----------------------------\n")
	return nil
}

func _main(args []string, env envConfig) int {
	ctx := context.Background()

	t, err := cloudeventsnats.New(env.ClusterID, env.ClientID, env.Subject)
	if err != nil {
		log.Fatalf("failed to create nats transport, %s", err.Error())
	}
	c, err := client.New(t)
	if err != nil {
		log.Fatalf("failed to create client, %s", err.Error())
	}

	if err := c.StartReceiver(ctx, receive); err != nil {
		log.Fatalf("failed to start nats receiver, %s", err.Error())
	}

	// Wait until done.
	<-ctx.Done()
	return 0
}
