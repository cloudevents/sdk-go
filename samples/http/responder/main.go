/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type envConfig struct {
	// Port on which to listen for cloudevents
	Port int    `envconfig:"RCV_PORT" default:"8080"`
	Path string `envconfig:"RCV_PATH" default:"/"`
}

type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func eventReceiver(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	fmt.Printf("Got Event Context: %+v\n", event.Context)
	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}
	fmt.Printf("Got Data: %+v\n", data)
	fmt.Printf("----------------------------\n")

	if data.Sequence%3 == 0 {
		responseEvent := cloudevents.NewEvent()
		responseEvent.SetID(uuid.New().String())
		responseEvent.SetSource("/mod3")
		responseEvent.SetType("samples.http.mod3")
		responseEvent.SetSubject(fmt.Sprintf("%s#%d", event.Source(), data.Sequence))

		_ = responseEvent.SetData(cloudevents.ApplicationJSON, Example{
			Sequence: data.Sequence,
			Message:  "mod 3!",
		})
		return &responseEvent, nil
	} else if data.Sequence%7 == 0 {
		responseEvent := cloudevents.NewEvent()
		responseEvent.SetID(uuid.New().String())
		responseEvent.SetSource("/mod7")
		responseEvent.SetType("samples.http.mod7")
		responseEvent.SetSubject(fmt.Sprintf("%s#%d", event.Source(), data.Sequence))

		_ = responseEvent.SetData(cloudevents.ApplicationJSON, Example{
			Sequence: data.Sequence,
			Message:  "mod 7 has issues!",
		})
		return &responseEvent, cloudevents.NewHTTPResult(500, "this is a mod 7 server error message")
	}

	return nil, nil
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}

	ctx := context.Background()

	p, err := cloudevents.NewHTTP(cloudevents.WithPort(env.Port), cloudevents.WithPath(env.Path))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}
	c, err := cloudevents.NewClient(p,
		cloudevents.WithUUIDs(),
		cloudevents.WithTimeNow(),
	)
	if err != nil {
		log.Fatalf("failed to create client: %s", err.Error())
	}

	if err := c.StartReceiver(ctx, eventReceiver); err != nil {
		log.Fatalf("failed to start receiver: %s", err.Error())
	}

	log.Printf("listening on :%d%s\n", env.Port, env.Path)
	<-ctx.Done()
}
