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

	receiver, err := kafka_sarama.NewConsumer([]string{"127.0.0.1:9092"}, saramaConfig, "test-group-id", "test-topic")
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(receiver)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("will listen consuming topic test-topic\n")
	err = c.StartReceiver(context.TODO(), receive)
	if err != nil {
		log.Fatalf("failed to start receiver: %s", err)
	} else {
		log.Printf("receiver stopped\n")
	}
}

func receive(ctx context.Context, event cloudevents.Event) {
	fmt.Printf("%s", event)
}
