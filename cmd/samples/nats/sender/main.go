package main

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventsnats "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"log"
	"net/url"
	"os"
	"time"
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

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	os.Exit(_main(os.Args[1:], env))
}

// Simple holder for the sending sample.
type Demo struct {
	Message string
	Source  url.URL
	Target  url.URL

	Client client.Client
}

// Basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func (d *Demo) Send(eventContext cloudevents.EventContext, i int) (*cloudevents.Event, error) {
	event := cloudevents.Event{
		Context: eventContext,
		Data: &Example{
			Sequence: i,
			Message:  d.Message,
		},
	}
	return d.Client.Send(context.Background(), event)
}

func _main(args []string, env envConfig) int {
	source, err := url.Parse("https://github.com/cloudevents/sdk-go/cmd/samples/sender")
	if err != nil {
		log.Printf("failed to parse source url, %v", err)
		return 1
	}

	seq := 0
	for _, contentType := range []string{"application/json", "application/xml"} {
		t, err := cloudeventsnats.New(env.NATSServer, env.Subject)
		if err != nil {
			log.Printf("failed to create nats transport, %s", err.Error())
			return 1
		}
		c, err := client.New(t)
		if err != nil {
			log.Printf("failed to create client, %s", err.Error())
			return 1
		}

		d := &Demo{
			Message: fmt.Sprintf("Hello, %s!", contentType),
			Source:  *source,
			Client:  c,
		}

		for i := 0; i < count; i++ {
			now := time.Now()
			ctx := cloudevents.EventContextV01{
				EventID:     uuid.New().String(),
				EventType:   "com.cloudevents.sample.sent",
				EventTime:   &types.Timestamp{Time: now},
				Source:      types.URLRef{URL: d.Source},
				ContentType: &contentType,
			}.AsV01()
			if _, err := d.Send(ctx, seq); err != nil {
				log.Printf("failed to send: %v", err)
				return 1
			}
			seq++
			time.Sleep(100 * time.Millisecond)
		}
	}

	return 0
}
