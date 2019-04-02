package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

const (
	count = 1000
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
	source, err := url.Parse("https://github.com/cloudevents/sdk-go/cmd/samples/sender")
	if err != nil {
		log.Printf("failed to parse source url, %v", err)
		return 1
	}

	seq := 0
	for _, contentType := range []string{"application/json", "application/xml"} {
		for _, encoding := range []cloudevents.HTTPEncoding{cloudevents.HTTPBinaryV01, cloudevents.HTTPStructuredV01, cloudevents.HTTPBinaryV02, cloudevents.HTTPStructuredV02} {

			t, err := cloudevents.NewHTTPTransport(
				cloudevents.WithTarget(env.Target),
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

			message := fmt.Sprintf("Hello, %s!", encoding)

			for i := 0; i < count; i++ {
				event := cloudevents.Event{
					Context: cloudevents.EventContextV01{
						EventID:     uuid.New().String(),
						EventType:   "com.cloudevents.sample.sent",
						Source:      cloudevents.URLRef{URL: *source},
						ContentType: &contentType,
					}.AsV01(),
					Data: &Example{
						Sequence: i,
						Message:  message,
					},
				}

				if resp, err := c.Send(context.Background(), event); err != nil {
					log.Printf("failed to send: %v", err)
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
