package main

import (
	"context"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.uber.org/atomic"
	"log"
	"os"
	"time"
)

var count = atomic.Int32{}

func main() {

	ctx := context.Background()
	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	go func() {
		t := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-t.C:
				_, _ = fmt.Fprintf(os.Stderr, "%d\n", count.Swap(0))
			case <-ctx.Done():
				return
			}
		}
	}()

	log.Printf("will listen on :8080\n")
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, receive))
}

func receive(ctx context.Context, event cloudevents.Event) {
	//fmt.Printf("%s", event)
	count.Inc()
}
