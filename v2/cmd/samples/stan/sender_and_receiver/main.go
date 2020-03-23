package main

import (
	"context"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	ce_stan "github.com/cloudevents/sdk-go/v2/protocol/stan"
	"log"
)

func main() {
	protocol, err := ce_stan.NewProtocol("test-cluster", "test-client", "send-test-subject", "receiver-test-subject",
		ce_stan.StanOptions())
	if err != nil {
		log.Fatalf("failed to create protocol: %v", err)
	}

	c, err := cloudevents.NewClient(protocol, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// Start the receiver
	go func() {
		log.Printf("will listen consuming topic test-topic\n")
		err = c.StartReceiver(context.TODO(), receive)
		if err != nil {
			log.Fatalf("failed to start receiver: %s", err)
		} else {
			log.Printf("receiver stopped\n")
		}
	}()

	// Start sending the events
	for i := 0; i < 10; i++ {
		e := cloudevents.NewEvent()
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/v2/cmd/samples/stan/sender_and_receiver")
		_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
			"id":      i,
			"message": "Hello, World!",
		})

		err := c.Send(context.Background(), e)
		if err != nil {
			log.Printf("failed to send: %v", err)
		} else {
			log.Printf("sent: %d", i)
		}
	}
}

func receive(ctx context.Context, event cloudevents.Event) {
	fmt.Printf("%s", event)
}
