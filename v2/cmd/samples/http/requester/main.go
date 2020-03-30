package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

const (
	count = 1
)

type envConfig struct {
	// Target URL where to send cloudevents
	Target string `envconfig:"TARGET" default:"http://localhost:8080" required:"true"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	os.Exit(_main(os.Args[1:], env))
}

// Basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func _main(args []string, env envConfig) int {
	source, err := url.Parse("https://github.com/cloudevents/sdk-go/v2/cmd/samples/requester")
	if err != nil {
		log.Printf("failed to parse source url, %v", err)
		return 1
	}

	seq := 0
	for _, contentType := range []string{"application/json", "application/xml", "text/plain"} {
		for _, encoding := range []cloudevents.Encoding{cloudevents.EncodingBinary, cloudevents.EncodingStructured} {

			p, err := cloudevents.NewHTTP(cloudevents.WithTarget(env.Target))
			if err != nil {
				log.Printf("failed to create protocol, %v", err)
				return 1
			}

			c, err := cloudevents.NewClient(p,
				cloudevents.WithTimeNow(),
			)
			if err != nil {
				log.Printf("failed to create client, %v", err)
				return 1
			}

			enc := "binary"
			if encoding == cloudevents.EncodingStructured {
				enc = "structured"
			}
			message := fmt.Sprintf("Hello %s, %s!", contentType, enc)

			for i := 0; i < count; i++ {
				event := cloudevents.Event{
					Context: cloudevents.EventContextV1{
						ID:     uuid.New().String(),
						Type:   "com.cloudevents.sample.sent",
						Source: cloudevents.URIRef{URL: *source},
					}.AsV1(),
				}
				_ = event.SetData(contentType, &Example{
					Sequence: i,
					Message:  message,
				})

				ctx := context.Background()

				switch encoding {
				case cloudevents.EncodingBinary:
					ctx = cloudevents.WithEncodingBinary(ctx)
				case cloudevents.EncodingStructured:
					ctx = cloudevents.WithEncodingStructured(ctx)
				}

				if resp, err := c.Request(ctx, event); err != nil {
					log.Printf("failed to request: %v", err)
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
					log.Printf("event sent at %s", time.Now())
				}

				seq++
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	return 0
}
