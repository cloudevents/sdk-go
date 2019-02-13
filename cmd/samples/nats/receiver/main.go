package main

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/go-nats"
	"log"
	"os"
	"time"
)

type envConfig struct {
	// NatsServer URL to connect to the nats server.
	NatsServer string `envconfig:"NATS_SERVER" default:"http://localhost:4222" required:"true"`

	// Subject is the nats subject to subscribe for cloudevents on.
	Subject string `envconfig:"SUBJECT" default:"sample" required:"true"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	os.Exit(_main(os.Args[1:], env))
}

type Receiver struct{}

type Example struct {
	Sequence int    `json:"id"`
	Message  string `json:"message"`
}

func (r *Receiver) Receive(event cloudevents.Event) {
	fmt.Printf("Got Event Context: %+v\n", event.Context)

	data := &Example{}
	if err := event.DataAs(data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}
	fmt.Printf("Got Data: %+v\n", data)

	fmt.Printf("----------------------------\n")
}

func _main(args []string, env envConfig) int {
	//r := &Receiver{}
	//t := &cloudeventshttp.Transport{Receiver: r}

	//log.Printf("listening on port %s\n", env.Port)
	//log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", env.Port), t))

	// Connect to a server
	nc, _ := nats.Connect(env.NatsServer)

	// Simple Async Subscriber
	sub, err := nc.Subscribe(env.Subject, func(m *nats.Msg) {
		fmt.Printf("Received a message: %s\n", string(m.Data))
	})
	if err != nil {
		return 1
	}

	time.Sleep(5 * time.Minute)

	// Unsubscribe
	sub.Unsubscribe()

	// Drain
	//sub.Drain()

	//// Requests
	//msg, err := nc.Request("help", []byte("help me"), 10*time.Millisecond)
	//
	//// Replies
	//nc.Subscribe("help", func(m *Msg) {
	//	nc.Publish(m.Reply, []byte("I can help!"))
	//})

	// Drain connection (Preferred for responders)
	// Close() not needed if this is called.
	//nc.Drain()

	// Close connection
	nc.Close()

	//	return nil

	return 0
}
