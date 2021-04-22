/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

const (
	count = 10
)

type envConfig struct {
	// Target URL where to send cloudevents
	Target string `envconfig:"TARGET" default:"http://localhost:8080" required:"true"`
}

// Example is a basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}

	p, err := cloudevents.NewHTTP(cloudevents.WithTarget(env.Target))
	if err != nil {
		log.Fatalf("Failed to create protocol, %v", err)
	}

	c, err := cloudevents.NewClient(p,
		cloudevents.WithTimeNow(),
	)
	if err != nil {
		log.Fatalf("Failed to create client, %v", err)
	}

	// Iterate over content types and encodings
	seq := 0
	for _, contentType := range []string{"application/json", "application/xml"} {
		for _, encoding := range []cloudevents.Encoding{cloudevents.EncodingBinary, cloudevents.EncodingStructured} {
			for i := 0; i < count; i++ {
				// Create a new event
				event := cloudevents.NewEvent()
				event.SetID(uuid.New().String())
				event.SetType("com.cloudevents.sample.sent")
				event.SetSource("https://github.com/cloudevents/sdk-go/v2/samples/requester")

				// Set data
				message := fmt.Sprintf("Sending message with content-type '%s' and encoding '%s'", contentType, encoding.String())
				_ = event.SetData(contentType, &Example{
					Sequence: i,
					Message:  message,
				})

				// Decorate the context forcing the encoding
				ctx := context.Background()
				switch encoding {
				case cloudevents.EncodingBinary:
					ctx = cloudevents.WithEncodingBinary(ctx)
				case cloudevents.EncodingStructured:
					ctx = cloudevents.WithEncodingStructured(ctx)
				}

				resp, result := c.Request(ctx, event)
				if cloudevents.IsUndelivered(result) {
					log.Printf("Failed to deliver request: %v", result)
				} else {
					// Event was delivered, but possibly not accepted and without a response.
					log.Printf("Event delivered at %s, Acknowledged==%t ", time.Now(), cloudevents.IsACK(result))
					var httpResult *cehttp.Result
					if cloudevents.ResultAs(result, &httpResult) {
						log.Printf("Response status code %d", httpResult.StatusCode)
					}
					// Request can get a response of nil, which is ok.
					if resp != nil {
						fmt.Printf("Response,\n%s\n", resp)
						fmt.Printf("----------------------------\n")
					}
				}
				seq++
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}
