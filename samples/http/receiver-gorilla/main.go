/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	ctx := context.Background()
	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	h, err := cloudevents.NewHTTPReceiveHandler(ctx, p, receive)
	if err != nil {
		log.Fatalf("failed to create handler: %s", err.Error())
	}

	// Use a gorilla mux implementation for the overall http handler.
	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// an example API handler
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	router.Handle("/", h)

	log.Printf("will listen on :8080\n")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("unable to start http server, %s", err)
	}
}

func receive(ctx context.Context, event cloudevents.Event) {
	fmt.Printf("Got an Event: %s", event)
}
