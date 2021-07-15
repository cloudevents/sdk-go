/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"

	cloudeventsjsm "github.com/cloudevents/sdk-go/protocol/nats_jetstream/v2"
	cloudeventshttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

type envConfig struct {
	// Port on which to listen for cloudevents
	Port int `envconfig:"PORT" default:"8080"`

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
	ctx := context.Background()

	natsProtocol, err := cloudeventsjsm.NewSender(env.NATSServer, "ORDER", env.Subject, cloudeventsjsm.NatsOptions(), nil)
	if err != nil {
		log.Fatalf("failed to create nats protcol, %s", err.Error())
	}

	defer natsProtocol.Close(ctx)

	httpProtocol, err := cloudeventshttp.New(cloudeventshttp.WithPort(env.Port))
	if err != nil {
		log.Fatalf("failed to create http protocol: %s", err.Error())
	}

	// Pipe all messages incoming to the httpProtocol to the natsProtocol
	go func() {
		for {
			// Blocking call to wait for new messages from httpProtocol
			message, err := httpProtocol.Receive(ctx)
			if err != nil {
				if err == io.EOF {
					return // Context closed and/or receiver closed
				}
				log.Printf("Error while receiving a message: %s", err.Error())
			}
			// Send message directly to natsProtocol
			err = natsProtocol.Send(ctx, message)
			if err != nil {
				log.Printf("Error while forwarding the message: %s", err.Error())
			}
		}
	}()

	// Start the HTTP Server invoking OpenInbound()
	go func() {
		if err := httpProtocol.OpenInbound(ctx); err != nil {
			log.Printf("failed to StartHTTPReceiver, %v", err)
		}
	}()

	<-ctx.Done()
}
