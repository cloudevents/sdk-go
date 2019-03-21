package main

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client/receiver"
	"log"
	"time"
)

func main() {
	ctx := context.Background()

	c, err := client.NewDefault()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("will listen on :8080\n")

	r := receiver.TypedReceiver{}
	if err := r.Add("io.cloudevents.colors.red", gotRed); err != nil {
		log.Printf("failed to register gotRed: %s", err.Error())
	}
	if err := r.Add("io.cloudevents.colors.blue", gotBlue); err != nil {
		log.Printf("failed to register gotBlue: %s", err.Error())
	}
	if err := r.Add("io.cloudevents.colors.green", gotGreen); err != nil {
		log.Printf("failed to register gotGreen: %s", err.Error())
	}

	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			out()
		}
	}()

	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, r.Receive))
}

var red = 0
var blue = 0
var green = 0

func gotRed() {
	red++
}

func gotBlue() {
	blue++
}

func gotGreen() {
	green++
}

func out() {
	fmt.Printf("Red: %d, Blue: %d, Green: %d, Total: %d\n", red, blue, green, red+blue+green)
}
