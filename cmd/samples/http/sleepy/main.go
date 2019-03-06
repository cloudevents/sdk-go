package main

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"github.com/kelseyhightower/envconfig"
	"log"
	"os"
	"time"
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

func gotEvent(event cloudevents.Event) {
	fmt.Printf("Got Event Context: %+v\n", event.Context)
	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}
	fmt.Printf("Got Data: %+v\n", data)
	fmt.Printf("----------------------------\n")
}

func _main(args []string, env envConfig) int {
	ctx := context.Background()

	_, c, err := client.StartHTTPReceiver(ctx, gotEvent,
		client.WithHTTPPort(env.Port),
		client.WithHTTPPath(env.Path),
	)

	if err != nil {
		log.Fatalf("failed to start receiver: %s", err.Error())
	}

	log.Printf("listening on :%d%s\n", env.Port, env.Path)

	for {
		time.Sleep(5 * time.Second)
		if err := c.StopReceiver(ctx); err != nil {
			log.Fatalf("failed to stop receiver: %s", err.Error())
		}
		log.Printf("stopped @ %s", time.Now())

		time.Sleep(5 * time.Second)
		if _, err := c.StartReceiver(ctx, gotEvent); err != nil {
			log.Fatalf("failed to start receiver: %s", err.Error())
		}
		log.Printf("started @ %s", time.Now())
	}
}
