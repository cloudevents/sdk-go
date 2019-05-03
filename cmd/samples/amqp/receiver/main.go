package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/amqp"
	"github.com/kelseyhightower/envconfig"
	qp "pack.ag/amqp"
)

type envConfig struct {
	// AMQPServer URL to connect to the amqp server.
	AMQPServer string `envconfig:"AMQP_SERVER" default:"amqp://localhost:5672/" required:"true"`

	// Queue is the amqp queue name to interact with.
	Queue string `envconfig:"AMQP_QUEUE"`

	AccessKeyName string `envconfig:"AMQP_ACCESS_KEY_NAME" default:"guest"`
	AccessKey     string `envconfig:"AMQP_ACCESS_KEY" default:"password"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	os.Exit(_main(os.Args[1:], env))
}

type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func receive(ctx context.Context, event cloudevents.Event, resp *cloudevents.EventResponse) error {
	fmt.Printf("Got CloudEvent,\n%+v\n", event)
	fmt.Println("----------------------------")
	return nil
}

func _main(args []string, env envConfig) int {
	ctx := context.Background()

	t, err := amqp.New(env.AMQPServer, env.Queue,
		amqp.WithConnOpt(qp.ConnSASLPlain(env.AccessKeyName, env.AccessKey)),
	)
	if err != nil {
		log.Fatalf("failed to create amqp transport, %s", err.Error())
	}
	c, err := client.New(t)
	if err != nil {
		log.Fatalf("failed to create client, %s", err.Error())
	}

	if err := c.StartReceiver(ctx, receive); err != nil {
		log.Fatalf("failed to start amqp receiver, %s", err.Error())
	}

	// Wait until done.
	<-ctx.Done()
	return 0
}
