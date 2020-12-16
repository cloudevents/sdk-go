package main

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"

	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
)

// In order to run this test, look at documentation in https://github.com/cloudevents/sdk-go/blob/master/v2/samples/kafka/README.md
func main() {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_0_0_0

	ctx := context.Background()

	// With NewProtocol you can use the same client both to send and receive.
	protocol, err := kafka_sarama.NewProtocol([]string{"127.0.0.1:9092"}, saramaConfig, "output-topic", "input-topic")
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	defer protocol.Close(context.Background())

	// Pipe all incoming message, eventually transforming them
	go func() {
		for {
			// Blocking call to wait for new messages from protocol
			inputMessage, err := protocol.Receive(ctx)
			if err != nil {
				if err == io.EOF {
					return // Context closed and/or receiver closed
				}
				log.Printf("Error while receiving a inputMessage: %s", err.Error())
				continue
			}
			defer inputMessage.Finish(nil)

			outputMessage := inputMessage

			// If encoding is unknown, then the inputMessage is a non cloudevent
			// and we need to convert it
			if inputMessage.ReadEncoding() == binding.EncodingUnknown {
				// We need to get the inputMessage internals
				// Because the message could be wrapped by the protocol implementation
				// we need to unwrap it and then cast to the message representation
				// specific to the protocol
				kafkaMessage := binding.UnwrapMessage(inputMessage).(*kafka_sarama.Message)

				// Now let's create a new event
				event := cloudevents.NewEvent()
				event.SetID(uuid.New().String())
				event.SetTime(time.Now())
				event.SetType("generated.examples")
				event.SetSource("https://github.com/cloudevents/sdk-go/v2/samples/kafka/sender")

				err = event.SetData(kafkaMessage.ContentType, kafkaMessage.Value)
				if err != nil {
					log.Printf("Error while setting the event data: %s", err.Error())
					continue
				}
				outputMessage = binding.ToMessage(&event)
			}

			// Send outputMessage directly to output-topic
			err = protocol.Send(ctx, outputMessage)
			if err != nil {
				log.Printf("Error while forwarding the inputMessage: %s", err.Error())
			}
		}
	}()

	// Start the Kafka Consumer Group invoking OpenInbound()
	go func() {
		if err := protocol.OpenInbound(ctx); err != nil {
			log.Printf("failed to StartHTTPReceiver, %v", err)
		}
	}()

	<-ctx.Done()
}
