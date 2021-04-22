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
)

func main() {
	ctx := context.Background()
	p, err := cloudevents.NewHTTP(
		cloudevents.WithDefaultOptionsHandlerFunc([]string{"POST", "OPTIONS"}, 100, []string{"http://localhost:8181"}, true),
	)
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
	fmt.Printf("%s", event)
}

//
// Testing with:
//
// cd ./tools; PORT=8181 go run ./http/raw/
//
// curl http://localhost:8080 -v -X OPTIONS -H "Origin: http://example.com" -H "WebHook-Request-Origin: http://example.com" -H "WebHook-Request-Callback: http://localhost:8181/do-this?now=true"
//
