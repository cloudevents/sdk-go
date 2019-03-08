package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	cloudeventsnats "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	// Port on which to listen for cloudevents
	Port int `envconfig:"PORT" default:"8080"`

	// NATSServer URL to connect to the nats server.
	NATSServer string `envconfig:"NATS_SERVER" default:"http://localhost:4222" required:"true"`

	// Subject is the nats subject to publish cloudevents on.
	Subject string `envconfig:"SUBJECT" default:"sample" required:"true"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	os.Exit(_main(os.Args[1:], env))
}

type Receiver struct {
	Client client.Client
}

type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func (r *Receiver) Receive(event cloudevents.Event) error {
	fmt.Printf("Got Event Context: %+v\n", event.Context)

	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}
	fmt.Printf("Got Data: %+v\n", data)

	fmt.Printf("forwarding...")

	if _, err := r.Client.Send(context.Background(), event); err != nil {
		fmt.Printf("forwarding failed: %s", err.Error())
	}

	fmt.Printf("----------------------------\n")
	return nil
}

func _main(args []string, env envConfig) int {
	ctx := context.Background()

	nt, err := cloudeventsnats.New(env.NATSServer, env.Subject)
	if err != nil {
		log.Fatalf("failed to create nats transport, %s", err.Error())
	}
	nc, err := client.New(nt)
	if err != nil {
		log.Printf("failed to create client, %v", err)
		return 1
	}

	r := &Receiver{Client: nc}

	t, err := cloudeventshttp.New(
		cloudeventshttp.WithPort(env.Port),
	)
	if err != nil {
		log.Printf("failed to create transport, %v", err)
		return 1
	}
	c, err := client.New(t)
	if err != nil {
		log.Printf("failed to create client, %v", err)
		return 1
	}
	err = c.StartReceiver(ctx, r.Receive)

	if err != nil {
		log.Printf("failed to StartHTTPReceiver, %v", err)
	}

	log.Printf("listening on port %d\n", env.Port)
	<-ctx.Done()

	return 0
}
