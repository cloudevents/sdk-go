package main

// A simple example of a source that produces multiple types of events.

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"log"
	"math/rand"
	"time"
)

var source = types.ParseURLRef("https://github.com/cloudevents/sdk-go/cmd/samples/simple/http/colors")

var eventTypes = []string{
	"io.cloudevents.colors.red",
	"io.cloudevents.colors.blue",
	"io.cloudevents.colors.green",
}

// Basic data struct.
type example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func main() {
	ctx := cecontext.WithTarget(context.Background(), "http://localhost:8080/")

	c, err := client.NewDefault()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	for i := 0; i < 1000; i++ {
		data := &example{
			Sequence: i,
			Message:  "Hello, Color World!",
		}

		event := cloudevents.Event{
			Context: cloudevents.EventContextV02{
				Type:   eventTypes[rand.Intn(len(eventTypes))],
				Source: *source,
			}.AsV02(),
			Data: data,
		}

		if resp, err := c.Send(ctx, event); err != nil {
			log.Printf("failed to send: %v", err)
		} else if resp != nil {
			fmt.Printf("got back a response: \n%s", resp)
		} else {
			log.Printf("%s: %d - %s", event.Type(), data.Sequence, data.Message)
		}

		time.Sleep(time.Millisecond * 50) // demo in human time.
	}
}
