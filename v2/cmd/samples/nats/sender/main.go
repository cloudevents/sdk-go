package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
	cloudeventsnats "github.com/cloudevents/sdk-go/v2/protocol/nats"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

const (
	count = 10
)

type envConfig struct {
	// NATSServer URL to connect to the nats server.
	NATSServer string `envconfig:"NATS_SERVER" default:"http://localhost:4222" required:"true"`

	// Subject is the nats subject to publish cloudevents on.
	Subject string `envconfig:"SUBJECT" default:"sample" required:"true"`
}

type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}

	for _, contentType := range []string{"application/json", "application/xml"} {
		p, err := cloudeventsnats.NewSender(env.NATSServer, env.Subject, cloudeventsnats.NatsOptions())
		if err != nil {
			log.Printf("failed to create nats protocol, %s", err.Error())
			os.Exit(1)
		}
		c, err := client.New(p)
		if err != nil {
			log.Printf("failed to create client, %s", err.Error())
			os.Exit(1)
		}

		for i := 0; i < count; i++ {
			now := time.Now()
			e := event.New()
			e.SetID(uuid.New().String())

			e.SetType("com.cloudevents.sample.sent")
			e.SetTime(now)
			e.SetSource("https://github.com/cloudevents/sdk-go/v2/cmd/samples/sender")
			_ = e.SetData(contentType, &Example{
				Sequence: i,
				Message:  fmt.Sprintf("Hello, %s!", contentType),
			})

			if err := c.Send(context.Background(), e); err != nil {
				log.Printf("failed to send: %v", err)
				os.Exit(1)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	os.Exit(0)
}
