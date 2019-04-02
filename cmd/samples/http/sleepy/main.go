package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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

func gotEvent(event cloudevents.Event) error {
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

	if err != nil {
		log.Fatalf("failed to start receiver: %s", err.Error())
	}

	log.Printf("listening on :%d%s\n", env.Port, env.Path)

	for {
		ctx, cancel := context.WithCancel(context.TODO())

		go func() {
			time.Sleep(5 * time.Second)
			cancel()
		}()

		log.Printf("starting @ %s", time.Now())
		if err := c.StartReceiver(ctx, gotEvent); err != nil {
			log.Fatalf("failed to start receiver: %s", err.Error())
		}
		log.Printf("stopped @ %s", time.Now())
		time.Sleep(5 * time.Second)
	}
}
