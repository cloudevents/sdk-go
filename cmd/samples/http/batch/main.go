package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudevents/sdk-go"
)

var source = cloudevents.ParseURLRef("https://github.com/cloudevents/sdk-go/cmd/samples/sender")

// Basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func main() {
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	ctx = cloudevents.ContextWithHeader(ctx, "demo", "header value")

	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	events := []cloudevents.Event(nil)
	for i := 0; i < 10; i++ {
		data := &Example{
			Sequence: i,
			Message:  "Hello, World!",
		}
		events = append(events, cloudevents.Event{
			Context: cloudevents.EventContextV02{
				Type:   "com.cloudevents.sample.sent",
				Source: *source,
			}.AsV02(),
			Data: data,
		})
	}

	if resp, err := c.Send(ctx, events...); err != nil {
		log.Printf("failed to send: %v", err)
	} else if resp != nil {
		fmt.Printf("got back a response: \n%s", resp)
	} else {
		//log.Printf("%s: %d - %s", event.Context.GetType(), data.Sequence, data.Message)
	}

}

/*

Producer -A-> MW -B-> Consumer


Producer Sends: E1, resp: 200 ack.
Producer Sends: E2, resp: 200 ack.

A: MW Accepts and Queues E1, E2

B: MW Sends [E1, E2] to Consumer Batched

Consumer Client accepts [E1, E2], resp is batch: [], or [En', Em', ...], or 400,500 error

Internally Consumer Client SDK delivers 1 event at a time with a resp.


*/
