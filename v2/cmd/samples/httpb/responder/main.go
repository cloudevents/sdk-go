package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/cloudevents/sdk-go/v2/protocol"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	// Port on which to listen for cloudevents
	//Port int    `envconfig:"RCV_PORT" default:"8080"`
	//Path string `envconfig:"RCV_PATH" default:"/"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	ctx := context.Background()

	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(p,
		cloudevents.WithUUIDs(),
		cloudevents.WithTimeNow(),
	)
	if err != nil {
		log.Fatalf("failed to create client: %s", err.Error())
	}

	log.Printf("listening on :%d%s\n", p.GetPort(), p.GetPath())
	if err := c.StartReceiver(ctx, gotEvent); err != nil {
		log.Fatalf("failed to start receiver: %s", err.Error())
	}
}

func gotEvent(ctx context.Context, event cloudevents.Event) (*event.Event, protocol.Result) {
	fmt.Printf("Got Event: %+v\n", event)

	// 50% chance of ACK
	if rand.Int()%2 == 0 {
		fmt.Printf(" ACK\n")
		return &event, cloudevents.NewHTTPResult(http.StatusAccepted, "accept")
	}

	fmt.Printf("NACK\n")
	return nil, cloudevents.NewHTTPResult(http.StatusBadRequest, "rejected")
}
