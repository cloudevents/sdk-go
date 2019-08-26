package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	// Port on which to listen for cloudevents
	Port int    `envconfig:"RCV_PORT" default:"8080"`
	Path string `envconfig:"RCV_PATH" default:"/"`
}

// Basic data struct.
type Example struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

func gotEvent(ctx context.Context, event cloudevents.Event) error {
	fmt.Printf("CloudEvent.Event: %+v\n", event)
	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Data Error: %s\n", err.Error())
	}
	fmt.Printf("Structured Data: %+v\n", data)

	fmt.Printf("Transport Context: %+v\n", cloudevents.HTTPTransportContextFrom(ctx))

	fmt.Printf("----------------------------\n")
	return nil
}

func convert(ctx context.Context, m transport.Message, err error) (*cloudevents.Event, error) {
	log.Printf("asked to convert, %v", m)
	log.Printf("trying to recover from %v", err)

	if msg, ok := m.(*http.Message); ok {
		tx := cloudevents.HTTPTransportContextFrom(ctx)

		data := &Example{}
		if err := json.Unmarshal(msg.Body, data); err != nil {
			return nil, err
		}

		// Make a new event and convert the message payload.
		event := cloudevents.NewEvent()
		event.SetSource("github.com/cloudevents/cmd/samples/http/converter/receiver")
		event.SetType(fmt.Sprintf("io.cloudevents.converter.http.%s", strings.ToLower(tx.Method)))
		event.SetID(uuid.New().String())
		event.SetSubject(fmt.Sprintf("%d", data.ID))

		if err := event.SetData(data); err != nil {
			return nil, err
		}

		return &event, nil
	}
	return nil, err
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	ctx := context.Background()

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithPort(env.Port),
		cloudevents.WithPath(env.Path),
	)
	if err != nil {
		log.Printf("failed to create transport, %v", err)
		os.Exit(1)
	}
	c, err := cloudevents.NewClient(t, cloudevents.WithConverterFn(convert))
	if err != nil {
		log.Printf("failed to create client, %v", err)
		os.Exit(1)
	}

	log.Printf("will listen on :%d%s\n", env.Port, env.Path)
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, gotEvent))
}

/*

curl -X POST -H "Content-Type: application/json"  -d '{"id":123,"message":"hello world"}' http://localhost:8080

*/
