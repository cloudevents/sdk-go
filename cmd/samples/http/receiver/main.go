package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	// Port on which to listen for cloudevents
	Port int    `envconfig:"RCV_PORT" default:"8080"`
	Path string `envconfig:"RCV_PATH" default:"/"`
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

func gotEvent(ctx context.Context, event cloudevents.Event) error {
	fmt.Printf("Got Event Context: %+v\n", event.Context)
	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}
	fmt.Printf("Got Data: %+v\n", data)

	fmt.Printf("Got Transport Context: %+v\n", cloudevents.HTTPTransportContextFrom(ctx))

	fmt.Printf("----------------------------\n")
	return nil
}

func _main(args []string, env envConfig) int {
	ctx := context.Background()

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithPort(env.Port),
		cloudevents.WithPath(env.Path),
	)
	if err != nil {
		log.Printf("failed to create transport, %v", err)
		return 1
	}
	c, err := cloudevents.NewClient(t)
	if err != nil {
		log.Printf("failed to create client, %v", err)
		return 1
	}

	log.Printf("will listen on :%d%s\n", env.Port, env.Path)
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, gotEvent))

	return 0
}
