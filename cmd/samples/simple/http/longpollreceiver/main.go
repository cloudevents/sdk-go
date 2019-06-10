package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/google/uuid"
)

func main() {
	ctx := context.Background()

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithBinaryEncoding(),
		cloudevents.WithLongPollTarget("http://localhost:8181/"),
	)
	if err != nil {
		log.Fatalf("failed to create transport, %v", err)
	}
	c, err := cloudevents.NewClient(t,
		cloudevents.WithTimeNow(),
		cloudevents.WithUUIDs(),
		cloudevents.WithConverterFn(convert),
	)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("will listen on :8080\n")
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, gotEvent))
}

// Example is the expected incoming event.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func convert(ctx context.Context, m transport.Message, err error) (*cloudevents.Event, error) {
	log.Printf("trying to recover from %v", err)

	if msg, ok := m.(*http.Message); ok {
		// Make a new event and convert the message payload.
		event := cloudevents.NewEvent()
		event.SetSource("github.com/cloudevents/cmd/samples/http/receiver")
		event.SetType("io.cloudevents.converter.http")
		event.SetID(uuid.New().String())
		event.SetDataContentType(cloudevents.ApplicationJSON)
		event.Data = msg.Body
		// Note: could use the msg headers as extensions.
		return &event, nil
	}
	return nil, err
}

func gotEvent(ctx context.Context, event cloudevents.Event) {
	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("failed to get data as Example: %s\n", err.Error())
		return
	}

	fmt.Printf("%s", event)
	fmt.Printf("%s\n", cloudevents.HTTPTransportContextFrom(ctx))
}
