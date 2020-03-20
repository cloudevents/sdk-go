package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
	ceamqp "github.com/cloudevents/sdk-go/v2/protocol/amqp"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/google/uuid"
	"pack.ag/amqp"
)

const (
	count = 10
)

// Parse AMQP_URL env variable. Return server URL, AMQP node (from path) and SASLPlain
// option if user/pass are present.
func sampleConfig() (server, node string, opts []ceamqp.Option) {
	env := os.Getenv("AMQP_URL")
	if env == "" {
		env = "/test"
	}
	u, err := url.Parse(env)
	if err != nil {
		log.Fatal(err)
	}
	if u.User != nil {
		user := u.User.Username()
		pass, _ := u.User.Password()
		opts = append(opts, ceamqp.WithConnOpt(amqp.ConnSASLPlain(user, pass)))
	}
	return env, strings.TrimPrefix(u.Path, "/"), opts
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

func (d *Demo) Send(eventContext event.EventContext, i int) error {
	e := event.Event{
		Context: eventContext,
	}
	_ = e.SetData(cloudevents.ApplicationJSON,
		&Example{
			Sequence: i,
			Message:  d.Message,
		})
	return d.Client.Send(context.Background(), e)
}

func main() {
	host, node, opts := sampleConfig()
	t, err := ceamqp.NewProtocol(host, node, []amqp.ConnOption{}, []amqp.SessionOption{}, opts...)
	if err != nil {
		log.Fatalf("Failed to create amqp protocol: %v", err)
	}
	c, err := client.New(t)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Attributes for events
	source, _ := url.Parse("https://github.com/cloudevents/sdk-go/v2/cmd/samples/sender")
	contentType := "application/json"

	// Value for event data
	seq := 0

	d := &Demo{
		Message: fmt.Sprintf("Hello, %s!", contentType),
		Source:  *source,
		Client:  c,
	}

	for i := 0; i < count; i++ {
		now := time.Now()
		ctx := event.EventContextV03{
			ID:              uuid.New().String(),
			Type:            "com.cloudevents.sample.sent",
			Time:            &types.Timestamp{Time: now},
			Source:          types.URIRef{URL: d.Source},
			DataContentType: &contentType,
		}.AsV03()
		if err := d.Send(ctx, seq); err != nil {
			log.Fatalf("Failed to send: %v", err)
		}
		seq++
		time.Sleep(100 * time.Millisecond)
	}
}
