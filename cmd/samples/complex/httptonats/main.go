package main

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/kelseyhightower/envconfig"
	"log"
	"net/http"
	"os"
)

type envConfig struct {
	// Port on which to listen for cloudevents
	Port string `envconfig:"PORT" default:"8080"`

	// NatsServer URL to connect to the nats server.
	NatsServer string `envconfig:"NATS_SERVER" default:"http://localhost:4222" required:"true"`

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
	Client *client.Client
}

type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func (r *Receiver) Receive(event cloudevents.Event) {
	fmt.Printf("Got Event Context: %+v\n", event.Context)

	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}
	fmt.Printf("Got Data: %+v\n", data)

	fmt.Printf("forwarding...")

	if err := r.Client.Send(event); err != nil {
		fmt.Printf("forwarding failed: %s", err.Error())
	}

	fmt.Printf("----------------------------\n")
}

func _main(args []string, env envConfig) int {

	c, err := client.NewNatsClient(context.TODO(), env.NatsServer, env.Subject, 0)
	if err != nil {
		log.Printf("failed to create client, %v", err)
		return 1
	}

	t := &cloudeventshttp.Transport{
		Receiver: &Receiver{Client: c},
	}

	log.Printf("listening on port %s\n", env.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", env.Port), t))

	return 0
}
