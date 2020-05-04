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
	count = 1
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
				event.SetSource("https://github.com/cloudevents/sdk-go/v2/cmd/samples/requester")

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

				if resp, res := c.Request(ctx, event); !cloudevents.IsACK(res) {
					log.Printf("Failed to request: %v", res)
				} else if resp != nil {
					fmt.Printf("Response:\n%s\n", resp)
					fmt.Printf("Got Event Response Context: %+v\n", resp.Context)
					data := &Example{}
					if err := resp.DataAs(data); err != nil {
						fmt.Printf("Got Data Error: %s\n", err.Error())
					}
					fmt.Printf("Got Response Data: %+v\n", data)
					fmt.Printf("----------------------------\n")
				} else {
					// Parse result
					var httpResult *cehttp.Result
					cloudevents.ResultAs(res, &httpResult)
					log.Printf("Event sent at %s", time.Now())
					log.Printf("Response status code %d", httpResult.StatusCode)
				}

				seq++
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}
