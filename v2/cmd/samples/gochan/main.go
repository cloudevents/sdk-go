package main

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol/gochan"
	"log"
	"time"
)

func main() {
	c, err := cloudevents.NewClient(gochan.New(), cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Millisecond*50)) // wait

	// Start the receiver
	go func() {
		if err := c.StartReceiver(ctx, func(ctx context.Context, event cloudevents.Event) {
			log.Printf("[receiver] %s", event)
		}); err != nil && err.Error() != "context deadline exceeded" {
			log.Fatalf("[receiver] start receiver returned an error: %s", err)
		}
		log.Println("[receiver] stopped")
	}()

	// Start sending the events
	for i := 0; i < 10; i++ {
		e := cloudevents.NewEvent()
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/v2/cmd/samples/gochan")
		_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
			"id":      i,
			"message": "Hello, World!",
		})

		err := c.Send(ctx, e)
		if err != nil {
			log.Printf("[sender] failed to send: %v", err)
		} else {
			log.Printf("[sender] sent: %d", i)
		}
	}
	// Wait for the timeout.
	<-ctx.Done()
	cancel()
	log.Println("[sender] stopped")
}
