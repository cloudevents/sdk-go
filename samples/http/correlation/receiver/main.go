/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/extensions"
)

func main() {
	ctx := context.Background()
	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("will listen on :8080\n")
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, receive))
}

func receive(ctx context.Context, event cloudevents.Event) {
	fmt.Printf("Received Event:\n%s\n", event)

	// Extract the correlation extension
	if ext, ok := extensions.GetCorrelationExtension(event); ok {
		fmt.Printf("Correlation ID: %s\n", ext.CorrelationID)
		if ext.CausationID != "" {
			fmt.Printf("Causation ID: %s\n", ext.CausationID)
		}
	} else {
		fmt.Printf("No Correlation Extension found in event\n")
	}
	fmt.Println("-------------------------------------------------")
}
