/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"

	cejsm "github.com/cloudevents/sdk-go/protocol/nats_jetstream/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
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
		log.Fatalf("Failed to process env var: %s", err)
	}

	p, err := cejsm.NewSender(env.NATSServer, "ORDER", env.Subject, nil, nil)
	if err != nil {
		log.Fatalf("Failed to create nats protocol, %s", err.Error())
	}

	defer p.Close(context.Background())

	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("Failed to create client, %s", err.Error())
	}

	for _, contentType := range []string{"application/json", "application/xml"} {
		for i := 0; i < count; i++ {
			e := cloudevents.NewEvent()
			e.SetID(uuid.New().String())
			e.SetType("com.cloudevents.sample.sent")
			e.SetTime(time.Now())
			e.SetSource("https://github.com/cloudevents/sdk-go/v2/samples/sender")
			_ = e.SetData(contentType, &Example{
				Sequence: i,
				Message:  fmt.Sprintf("Hello, %s!", contentType),
			})

			if result := c.Send(context.Background(), e); cloudevents.IsUndelivered(result) {
				log.Printf("failed to send: %v", err)
			} else {
				log.Printf("sent: %d, accepted: %t", i, cloudevents.IsACK(result))
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
