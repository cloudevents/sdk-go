/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"log"

	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpcprotocol "github.com/cloudevents/sdk-go/protocol/grpc"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	ctx := context.Background()
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	p, err := grpcprotocol.NewProtocol(conn, grpcprotocol.WithSubscribeOption(&grpcprotocol.SubscribeOption{Topic: "test-topic"}))
	if err != nil {
		log.Fatalf("failed to create protocol: %v", err)
	}
	defer p.Close(ctx)

	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	log.Printf("Receiver start consuming messages from test-topic\n")
	err = c.StartReceiver(ctx, receive)
	if err != nil {
		log.Fatalf("failed to start receiver: %v", err)
	} else {
		log.Printf("receiver stopped")
	}
}

func receive(ctx context.Context, event cloudevents.Event) {
	log.Printf("received event:\n%s", event)
}
