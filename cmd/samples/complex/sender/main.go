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
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/transport"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/transport/http"
	cloudeventsnats "github.com/cloudevents/sdk-go/pkg/transport/nats"
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
}

var seq int

// Basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func (d *Demo) Send() error {
	e := cloudevents.Event{
		Context: d.context(),
	}
	_ = e.SetData(&Example{
		Sequence: seq,
		Message:  d.Message,
	}, cloudevents.ApplicationJSON)
	seq++
	return d.Client.Send(context.Background(), e)
}

func (d *Demo) context() cloudevents.EventContext {
	ctx := event.EventContextV1{
		Type:            d.EventType,
		Source:          types.URIRef{URL: d.Source},
		DataContentType: &d.ContentType,
	}.AsV1()
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
		for _, encoding := range []cloudeventshttp.Encoding{cloudeventshttp.Default, cloudeventshttp.Binary, cloudeventshttp.Structured} {

			p, err := cloudeventshttp.NewProtocol(cloudeventshttp.WithTarget(env.HTTPTarget))
			if err != nil {
				log.Printf("failed to create protocol, %v", err)
				return 1
			}

			t, err := cloudeventshttp.New(p,
				cloudeventshttp.WithEncoding(encoding),
			)
			if err != nil {
				log.Printf("failed to create transport, %v", err)
				return 1
			}

			if err := doDemo(
				t,
				"com.cloudevents.sample.http.sent",
				fmt.Sprintf("Hello %d, using %s!", encoding, contentType),
				contentType,
				*source,
			); err != nil {
				log.Printf("failed to do http demo: %v, %s", err, contentType)
				return 1
			}
		}

		// NATS
		t, err := cloudeventsnats.New(
			env.NATSServer,
			env.Subject,
		)
		if err != nil {
			log.Printf("failed to create client, %v", err)
			return 1
		}
		if err := doDemo(
			t.Transport(),
			"com.cloudevents.sample.nats.sent",
			fmt.Sprintf("Hello NATS, using %s!", contentType),
			contentType,
			*source,
		); err != nil {
			log.Printf("failed to do nats demo: %v, %s", err, contentType)
			return 1
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
		if err := d.Send(); err != nil {
			return err
		}
		time.Sleep(delay)
	}
	return nil
}
