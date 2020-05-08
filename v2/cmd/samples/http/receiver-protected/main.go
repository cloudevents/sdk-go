package main

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	ctx := context.Background()
	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	rate := 100
	p.WebhookConfig = &http.WebhookConfig{
		AllowedMethods:  []string{"POST", "OPTIONS"},
		AllowedRate:     &rate,
		AutoACKCallback: true,
		AllowedOrigins:  []string{"http://localhost:8181"},
	}
	p.OptionsHandlerFn = p.OptionsHandler

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
// PORT=8181 go run ./cmd/samples/http/raw/
//
// curl http://localhost:8080 -v -X OPTIONS -H "Origin: http://example.com" -H "WebHook-Request-Origin: http://example.com" -H "WebHook-Request-Callback: http://localhost:8181/do-this?now=true"
//
