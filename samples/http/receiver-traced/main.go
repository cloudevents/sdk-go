/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.opencensus.io/plugin/ochttp"
)

func main() {
	ctx := context.Background()
	p, err := cloudevents.NewHTTP(cloudevents.WithMiddleware(func(next http.Handler) http.Handler {
		return &ochttp.Handler{Handler: next}
	}))
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

func receive(ctx context.Context, e cloudevents.Event) {
	fmt.Printf("%s", e)
}
