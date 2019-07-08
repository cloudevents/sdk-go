package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/client"
	"github.com/cloudevents/sdk-go/pkg/transport"
	cehttp "github.com/cloudevents/sdk-go/pkg/transport/http"
	cenats "github.com/cloudevents/sdk-go/pkg/transport/nats"
	"github.com/cloudevents/sdk-go/pkg/types"
	"github.com/kelseyhightower/envconfig"
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

func (d *Demo) Send() (*cloudevents.Event, error) {
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
	ctx := cloudevents.EventContextV01{
		EventType:   d.EventType,
		Source:      types.URLRef{URL: d.Source},
		ContentType: &d.ContentType,
	}.AsV01()
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
		for _, encoding := range []cehttp.Encoding{cehttp.Default, cehttp.BinaryV01, cehttp.StructuredV01, cehttp.BinaryV02, cehttp.StructuredV02, cehttp.BinaryV03, cehttp.StructuredV03} {

			if err != nil {
				log.Printf("failed to create client, %v", err)
				return 1
			}

			t, err := cehttp.New(
				cehttp.WithTarget(env.HTTPTarget),
				cehttp.WithEncoding(encoding),
			)
			if err != nil {
				log.Printf("failed to create client, %v", err)
				return 1
			}

			if err := doDemo(
				t,
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
		for _, encoding := range []cenats.Encoding{cenats.Default, cenats.StructuredV02, cenats.StructuredV03} {

			t, err := cenats.New(
				env.NATSServer,
				env.Subject,
				cenats.WithEncoding(encoding),
			)
			if err != nil {
				log.Printf("failed to create client, %v", err)
				return 1
			}
			if err := doDemo(
				t,
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

func doDemo(t transport.Transport, eventType, message, contentType string, source url.URL) error {

	c, err := client.New(t,
		client.WithUUIDs(),
		client.WithTimeNow(),
	)
	if err != nil {
		return err
	}

	d := &Demo{
		Message:     message,
		Client:      c,
		Source:      source,
		EventType:   eventType,
		ContentType: contentType,
	}
	for i := 0; i < count; i++ {
		if _, err := d.Send(); err != nil {
			return err
		}
		time.Sleep(delay)
	}
	return nil
}
