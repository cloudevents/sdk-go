/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/google/uuid"

	ceamqp "github.com/cloudevents/sdk-go/protocol/amqp/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
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

// Example is a basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func main() {
	host, node, opts := sampleConfig()
	p, err := ceamqp.NewProtocol(host, node, []amqp.ConnOption{}, []amqp.SessionOption{}, opts...)
	if err != nil {
		log.Fatalf("Failed to create amqp protocol: %v", err)
	}

	// Close the connection when finished
	defer p.Close(context.Background())

	// Create a new client from the given protocol
	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	for i := 0; i < count; i++ {
		event := cloudevents.NewEvent()
		event.SetID(uuid.New().String())
		event.SetSource("https://github.com/cloudevents/sdk-go/v2/samples/sender")
		event.SetTime(time.Now())
		event.SetType("com.cloudevents.sample.sent")

		err := event.SetData(cloudevents.ApplicationJSON,
			&Example{
				Sequence: i,
				Message:  "Hello world!",
			})
		if err != nil {
			log.Fatalf("Failed to set data: %v", err)
		}

		if result := c.Send(context.Background(), event); cloudevents.IsUndelivered(result) {
			log.Fatalf("Failed to send: %v", result)
		} else if cloudevents.IsNACK(result) {
			log.Printf("Event not accepted: %v", result)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
