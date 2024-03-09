/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"
	"time"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"

	grpcprotocol "github.com/cloudevents/sdk-go/protocol/grpc"
)

const (
	count = 100
)

func main() {
	ctx := context.Background()
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	p, err := grpcprotocol.NewProtocol(conn, grpcprotocol.WithPublishOption(&grpcprotocol.PublishOption{Topic: "test-topic"}))
	if err != nil {
		log.Fatalf("failed to create protocol: %v", err)
	}
	defer p.Close(ctx)

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	for i := 0; i < count; i++ {
		e := cloudevents.NewEvent()
		e.SetID(uuid.New().String())
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/samples/grpc/sender")
		err = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
			"id":      i,
			"message": "Hello, World!",
		})
		if err != nil {
			log.Printf("failed to set data: %v", err)
		}
		if result := c.Send(ctx, e); cloudevents.IsUndelivered(result) {
			log.Printf("failed to send event: %v", result)
		} else {
			log.Printf("sent: %d, accepted: %t", i, cloudevents.IsACK(result))
		}
		time.Sleep(1 * time.Second)
	}
}
