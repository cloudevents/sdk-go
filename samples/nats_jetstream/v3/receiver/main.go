/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	cejsm "github.com/cloudevents/sdk-go/protocol/nats_jetstream/v3"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	ctx := context.Background()

	natsURL := "nats://localhost:4222"
	natsSubject := "sample.odd.>"
	natsStream := "stream"

	createStream(natsURL, natsStream, natsSubject)
	consumerOpt := cejsm.WithConsumerConfig(&jetstream.ConsumerConfig{FilterSubjects: []string{natsSubject}})
	urlOpt := cejsm.WithURL(natsURL)
	protocol, err := cejsm.New(ctx, consumerOpt, urlOpt)
	if err != nil {
		log.Fatalf("failed to create nats protocol: %s", err.Error())
	}

	defer protocol.Close(ctx)

	c, err := cloudevents.NewClient(protocol)
	if err != nil {
		log.Fatalf("failed to create client: %s", err.Error())
	}

	if err := c.StartReceiver(ctx, receive); err != nil {
		log.Printf("failed to start nats receiver: %s", err.Error())
	}
}

type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func receive(ctx context.Context, event cloudevents.Event) error {
	fmt.Printf("Got Event Context: %+v\n", event.Context)

	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}
	fmt.Printf("Got Data: %+v\n", data)

	fmt.Printf("----------------------------\n")
	return nil
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
