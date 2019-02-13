package main

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	cloudeventsnats "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/go-nats"
	"log"
	"os"
)

type envConfig struct {
	// NatsServer URL to connect to the nats server.
	NatsServer string `envconfig:"NATS_SERVER" default:"http://localhost:4222" required:"true"`

	// Subject is the nats subject to subscribe for cloudevents on.
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

type Receiver struct{}

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

	fmt.Printf("----------------------------\n")
}

func _main(args []string, env envConfig) int {

	conn, err := nats.Connect(env.NatsServer)
	if err != nil {
		log.Fatalf("failed to connect to nats server, %s", err.Error())
	}

	r := &Receiver{}
	t := &cloudeventsnats.Transport{
		Conn:     conn,
		Receiver: r,
	}

	ctx := context.TODO()

	err = t.Listen(ctx, env.Subject)
	if err != nil {
		log.Fatalf("failed to listen, %s", err.Error())
	}

	// Wait until done.
	<-ctx.Done()

	// Close connection.
	conn.Close()

	return 0
}
