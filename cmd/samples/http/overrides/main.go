package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/cloudevents/sdk-go"
	"github.com/kelseyhightower/envconfig"
)

const (
	count = 500
)

type envConfig struct {
	// Target URL where to send cloudevents
	Target    string `envconfig:"TARGET" default:"http://localhost:8080" required:"true"`
	Overrides string `envconfig:"OVERRIDES" default:""`
}

// Basic data struct.
type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}

	if env.Overrides == "" {
		_, filename, _, _ := runtime.Caller(0)
		dir := filepath.Dir(filename)
		env.Overrides = filepath.Join(dir, "extensions")
	}

	source := "https://github.com/cloudevents/sdk-go/cmd/samples/http/overrides"

	ctx := cloudevents.ContextWithEncoding(context.Background(), cloudevents.Structured)

	contentType := "application/json"

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(env.Target),
		cloudevents.WithBinaryEncoding(),
	)
	if err != nil {
		log.Printf("failed to create transport, %v", err)
		os.Exit(1)
	}

	c, err := cloudevents.NewClient(t,
		cloudevents.WithUUIDs(),
		cloudevents.WithTimeNow(),
		cloudevents.WithOverrides(ctx, env.Overrides),
	)
	if err != nil {
		log.Printf("failed to create client, %v", err)
		os.Exit(1)
	}

	message := fmt.Sprintf("Hello, %s!", "overrides")

	for i := 0; i < count; i++ {
		event := cloudevents.NewEvent("1.0")
		event.SetType("com.cloudevents.sample.overrides")
		event.SetSource(source)
		event.SetDataContentType(contentType)
		_ = event.SetData(&Example{
			Sequence: i,
			Message:  message,
		})

		_, _, _ = c.Send(ctx, event)

		time.Sleep(1 * time.Second)
	}

	os.Exit(0)
}
