/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"

	cestan "github.com/cloudevents/sdk-go/protocol/stan/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	receiver, err := cestan.NewConsumer("test-cluster", "test-client", "test-subject", cestan.StanOptions())
	if err != nil {
		log.Fatalf("failed to create protocol: %v", err)
	}

	defer receiver.Close(context.Background())

	c, err := cloudevents.NewClient(receiver, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	log.Printf("will listen consuming topic test-topic\n")
	err = c.StartReceiver(context.TODO(), receive)
	if err != nil {
		log.Fatalf("failed to start receiver: %s", err)
	} else {
		log.Printf("receiver stopped\n")
	}

}

func receive(_ context.Context, event cloudevents.Event) {
	fmt.Printf("%s", event)
}
