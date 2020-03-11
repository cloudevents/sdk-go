package main

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/event"
	"log"
	"os"

	"github.com/cloudevents/sdk-go/pkg/transport/http"

	cloudevents "github.com/cloudevents/sdk-go"
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

	p, err := cloudevents.NewHTTPProtocol()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	t, err := http.New(p)
	if err != nil {
		log.Fatalf("failed to create transport, %v", err)
	}

	c, err := cloudevents.NewClient(t,
		cloudevents.WithUUIDs(),
		cloudevents.WithTimeNow(),
	)
	if err != nil {
		log.Fatalf("failed to create client: %s", err.Error())
	}

	log.Printf("listening on :%d%s\n", t.GetPort(), t.GetPath())
	if err := c.StartReceiver(ctx, gotEvent); err != nil {
		log.Fatalf("failed to start receiver: %s", err.Error())
	}
}

func gotEvent(ctx context.Context, event cloudevents.Event, resp *cloudevents.EventResponse) event.Response {
	fmt.Printf("Got Event: %+v\n", event)

	resp.RespondWith(200, &event)
	return cloudevents.NewHTTPResponse(206, "accept")
}
