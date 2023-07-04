package main

import (
	"context"
	"fmt"
	"log"
	"net"

	cemqtt "github.com/cloudevents/sdk-go/protocol/mqtt_paho/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/eclipse/paho.golang/paho"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		log.Fatalf("failed to connect to mqtt broker: %s", err.Error())
	}
	clientConfig := &paho.ClientConfig{
		ClientID: "receiver-client-id",
		Conn:     conn,
	}
	cp := &paho.Connect{
		KeepAlive:  30,
		CleanStart: true,
	}

	p, err := cemqtt.New(context.TODO(), clientConfig, cp, "test-topic", []string{"test-topic"}, 0, false)
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}
	defer p.Close(context.Background())

	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("will listen consuming topic test-topic\n")
	err = c.StartReceiver(context.Background(), receive)
	if err != nil {
		log.Fatalf("failed to start receiver: %s", err)
	} else {
		log.Printf("receiver stopped\n")
	}
}

func receive(ctx context.Context, event cloudevents.Event) {
	fmt.Printf("%s", event)
}
