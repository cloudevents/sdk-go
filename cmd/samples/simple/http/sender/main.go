package main

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cecontext "github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"log"
)

var source = types.ParseURLRef("https://github.com/cloudevents/sdk-go/cmd/samples/sender")

// Basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func main() {
	ctx := cecontext.WithTarget(context.Background(), "http://localhost:8080/")

	ctx = cehttp.ContextWithHeader(ctx, "demo", "header value")

	c, err := client.NewDefault()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	for i := 0; i < 10; i++ {
		data := &Example{
			Sequence: i,
			Message:  "Hello, World!",
		}
		event := cloudevents.Event{
			Context: cloudevents.EventContextV02{
				Type:   "com.cloudevents.sample.sent",
				Source: *source,
			}.AsV02(),
			Data: data,
		}

		if resp, err := c.Send(ctx, event); err != nil {
			log.Printf("failed to send: %v", err)
		} else if resp != nil {
			fmt.Printf("got back a response event of type %s", resp.Context.GetType())
		} else {
			log.Printf("%s: %d - %s", event.Context.GetType(), data.Sequence, data.Message)
		}
	}
}
