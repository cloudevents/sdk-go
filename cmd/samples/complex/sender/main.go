package main

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
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
	count = 1
	delay = 100 * time.Millisecond
)

type envConfig struct {
	// HTTPTarget is the target URL where to send cloudevents
	HTTPTarget string `envconfig:"HTTP_TARGET" default:"http://localhost:8080" required:"true"`

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
	Client client.Client

	// Content
	EventType   string
	Source      url.URL
	ContentType string

	// Data
	Message string

	seq int
}

var seq int

// Basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func (d *Demo) Send() error {
	event := cloudevents.Event{
		Context: d.context(),
		Data: &Example{
			Sequence: seq,
			Message:  d.Message,
		},
	}
	seq++
	return d.Client.Send(context.Background(), event)
}

func (d *Demo) context() cloudevents.EventContext {
	now := time.Now()
	ctx := cloudevents.EventContextV01{
		EventID:     uuid.New().String(),
		EventType:   d.EventType,
		EventTime:   &types.Timestamp{Time: now},
		Source:      types.URLRef{URL: d.Source},
		ContentType: &d.ContentType,
	}
	return ctx
}

func _main(args []string, env envConfig) int {
	source, err := url.Parse("https://github.com/cloudevents/sdk-go/cmd/samples/sender")
	if err != nil {
		log.Printf("failed to parse source url, %v", err)
		return 1
	}

	for _, contentType := range []string{"application/json", "application/xml"} {
		// HTTP
		for _, encoding := range []cloudeventshttp.Encoding{cloudeventshttp.Default, cloudeventshttp.BinaryV01, cloudeventshttp.StructuredV01, cloudeventshttp.BinaryV02, cloudeventshttp.StructuredV02, cloudeventshttp.BinaryV03, cloudeventshttp.StructuredV03} {

			c, err := client.NewHTTPClient(client.WithTarget(env.HTTPTarget), client.WithHTTPEncoding(encoding))
			if err != nil {
				log.Printf("failed to create client, %v", err)
				return 1
			}

			if err := doDemo(
				c,
				"com.cloudevents.sample.http.sent",
				fmt.Sprintf("Hello, %s using %s!", encoding, contentType),
				contentType,
				*source,
			); err != nil {
				log.Printf("failed to do http demo: %v, %s", err, contentType)
				return 1
			}
		}

		// NATS
		for _, encoding := range []cloudeventsnats.Encoding{cloudeventsnats.Default, cloudeventsnats.StructuredV02, cloudeventsnats.StructuredV03} {
			c, err := client.NewNATSClient(env.NATSServer, env.Subject, client.WithNATSEncoding(encoding))
			if err != nil {
				log.Printf("failed to create client, %v", err)
				return 1
			}
			if err := doDemo(
				c,
				"com.cloudevents.sample.nats.sent",
				fmt.Sprintf("Hello, %s using %s!", encoding, contentType),
				contentType,
				*source,
			); err != nil {
				log.Printf("failed to do nats demo: %v, %s", err, contentType)
				return 1
			}
		}
	}

	return 0
}

func doDemo(c client.Client, eventType, message, contentType string, source url.URL) error {
	d := &Demo{
		Message:     message,
		Client:      c,
		Source:      source,
		EventType:   eventType,
		ContentType: contentType,
	}
	for i := 0; i < count; i++ {
		if err := d.Send(); err != nil {
			return err
		}
		time.Sleep(delay)
	}
	return nil
}
