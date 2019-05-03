package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/amqp"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	qp "pack.ag/amqp"
)

const (
	count = 10
)

type envConfig struct {
	// AMQPServer URL to connect to the amqp server.
	AMQPServer string `envconfig:"AMQP_SERVER" default:"amqp://localhost:5672/" required:"true"`

	// Queue is the amqp queue name to interact with.
	Queue string `envconfig:"AMQP_QUEUE"`

	AccessKeyName string `envconfig:"AMQP_ACCESS_KEY_NAME" default:"guest"`
	AccessKey     string `envconfig:"AMQP_ACCESS_KEY" default:"password"`
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
	contentType := "application/json"
	t, err := amqp.New(env.AMQPServer, env.Queue,
		amqp.WithConnOpt(qp.ConnSASLPlain(env.AccessKeyName, env.AccessKey)),
	)
	if err != nil {
		log.Printf("failed to create amqp transport, %s", err.Error())
		return 1
	}
	t.Encoding = amqp.BinaryV03
	//t.Encoding = amqp.StructuredV02
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
		ctx := cloudevents.EventContextV03{
			ID:              uuid.New().String(),
			Type:            "com.cloudevents.sample.sent",
			Time:            &types.Timestamp{Time: now},
			Source:          types.URLRef{URL: d.Source},
			DataContentType: &contentType,
		}.AsV03()
		if _, err := d.Send(ctx, seq); err != nil {
			log.Printf("failed to send: %v", err)
			return 1
		}
		seq++
		time.Sleep(100 * time.Millisecond)
	}

	return 0
}
