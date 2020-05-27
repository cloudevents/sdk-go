package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0

	receiver, err := kafka_sarama.NewConsumer([]string{"127.0.0.1:9092"}, saramaConfig, "test-group-id", "test-topic")
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	defer receiver.Close(context.Background())

	c, err := cloudevents.NewClient(receiver)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("will listen consuming topic test-topic\n")
	err = c.StartReceiver(context.Background(), receive)
	if err != nil {
		log.Fatalf("failed to start receiver: %s", err)
	} else {
		log.Printf("receiver stopped\n")
	}
}

func receive(ctx context.Context, event cloudevents.Event) {
	fmt.Printf("%s", event)
}
