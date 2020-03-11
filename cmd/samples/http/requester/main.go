package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
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
	source, err := url.Parse("https://github.com/cloudevents/sdk-go/cmd/samples/requester")
	if err != nil {
		log.Printf("failed to parse source url, %v", err)
		return 1
	}

	seq := 0
	for _, contentType := range []string{"application/json", "application/xml"} {
		for _, encoding := range []cloudevents.HTTPEncoding{cloudevents.HTTPBinaryEncoding, cloudevents.HTTPStructuredEncoding} {

			p, err := cloudevents.NewHTTPProtocol(cloudevents.WithTarget(env.Target))
			if err != nil {
				log.Printf("failed to create protocol, %v", err)
				return 1
			}

			t, err := cloudevents.NewHTTPTransport(p,
				cloudevents.WithEncoding(encoding),
			)
			if err != nil {
				log.Printf("failed to create transport, %v", err)
				return 1
			}

			c, err := cloudevents.NewClient(t,
				cloudevents.WithTimeNow(),
			)
			if err != nil {
				log.Printf("failed to create client, %v", err)
				return 1
			}

			message := fmt.Sprintf("Hello, %d!", encoding)

			for i := 0; i < count; i++ {
				event := cloudevents.Event{
					Context: cloudevents.EventContextV1{
						ID:              uuid.New().String(),
						Type:            "com.cloudevents.sample.sent",
						Source:          cloudevents.URIRef{URL: *source},
						DataContentType: &contentType,
					}.AsV1(),
					Data: &Example{
						Sequence: i,
						Message:  message,
					},
				}

				if resp, err := c.Request(context.Background(), event); err != nil {
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
