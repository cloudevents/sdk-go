package main

import (
	"context"
	"log"

	"github.com/Shopify/sarama"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/transport/kafka_sarama"
)

func main() {
	saramaConfig := sarama.NewConfig()

	sender, err := kafka_sarama.NewSender([]string{"127.0.0.1:9092"}, saramaConfig, "test-topic")
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(sender, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	for i := 0; i < 10; i++ {
		e := cloudevents.NewEvent()
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/cmd/samples/httpb/requester")
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
