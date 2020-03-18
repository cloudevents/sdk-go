package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Shopify/sarama"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol/kafka_sarama"
)

func main() {
	saramaConfig := sarama.NewConfig()

	// With NewProtocol you can use the same client both to send and receive.
	protocol, err := kafka_sarama.NewProtocol([]string{"127.0.0.1:9092"}, saramaConfig, "send-test-topic", "receive-test-topic")
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(protocol, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
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
		e.SetSource("https://github.com/cloudevents/sdk-go/v2/cmd/samples/httpb/requester")
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
