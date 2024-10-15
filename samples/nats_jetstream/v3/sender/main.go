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
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	cejsm "github.com/cloudevents/sdk-go/protocol/nats_jetstream/v3"
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

	natsURL := "nats://localhost:4222"
	natsSubject := "sample"
	natsStream := "stream"

	createStream(natsURL, natsStream, natsSubject)
	ctx := context.Background()
	urlOpt := cejsm.WithURL(natsURL)
	sendopt := cejsm.WithSendSubject(natsSubject)
	protocol, err := cejsm.New(ctx, urlOpt, sendopt)
	if err != nil {
		log.Fatalf("Failed to create nats protocol: %s", err.Error())
	}

	defer protocol.Close(context.Background())

	c, err := cloudevents.NewClient(protocol)
	if err != nil {
		log.Fatalf("Failed to create client: %s", err.Error())
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

func createStream(url, streamName, subjectName string) error {
	ctx := context.Background()
	streamConfig := jetstream.StreamConfig{Name: streamName, Subjects: []string{subjectName}}

	natsConn, err := nats.Connect(url)
	if err != nil {
		return err
	}
	js, err := jetstream.New(natsConn)
	if err != nil {
		return err
	}

	_, err = js.CreateOrUpdateStream(ctx, streamConfig)
	return err
}
