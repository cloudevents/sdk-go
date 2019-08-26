package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

func gotEvent(ctx context.Context, event cloudevents.Event, resp *cloudevents.EventResponse) error {
	fmt.Printf("Got Event Context: %+v\n", event.Context)
	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}
	fmt.Printf("Got Data: %+v\n", data)
	fmt.Printf("Got Transport Context: %+v\n", cloudevents.HTTPTransportContextFrom(ctx))
	fmt.Printf("----------------------------\n")

	if data.Sequence%3 == 0 {
		r := cloudevents.Event{
			Context: cloudevents.EventContextV02{
				Source: *cloudevents.ParseURLRef("/mod3"),
				Type:   "samples.http.mod3",
			}.AsV02(),
			Data: Example{
				Sequence: data.Sequence,
				Message:  "mod 3!",
			},
		}
		resp.RespondWith(200, &r)
		resp.Context = &cloudevents.HTTPTransportResponseContext{
			Header: func() http.Header {
				h := http.Header{}
				h.Set("sample", "magic header")
				h.Set("mod", "3")
				return h
			}(),
		}
		return nil
	}

	return nil
}

func _main(args []string, env envConfig) int {
	ctx := context.Background()

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithPort(env.Port),
		cloudevents.WithPath(env.Path),
	)
	if err != nil {
		log.Fatalf("failed to create transport: %s", err.Error())
	}
	c, err := cloudevents.NewClient(t,
		cloudevents.WithUUIDs(),
		cloudevents.WithTimeNow(),
	)
	if err != nil {
		log.Fatalf("failed to create client: %s", err.Error())
	}

	if err := c.StartReceiver(ctx, gotEvent); err != nil {
		log.Fatalf("failed to start receiver: %s", err.Error())
	}

	log.Printf("listening on :%d%s\n", env.Port, env.Path)
	<-ctx.Done()

	return 0
}
