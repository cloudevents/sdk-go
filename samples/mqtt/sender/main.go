package main

import (
	"context"
	"log"
	"net"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"

	cemqtt "github.com/cloudevents/sdk-go/protocol/mqtt_paho/v2"
)

const (
	count = 10
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		log.Fatalf("failed to connect to mqtt broker: %s", err.Error())
	}
	clientConfig := &paho.ClientConfig{
		ClientID: "sender-client-id",
		Conn:     conn,
	}
	cp := &paho.Connect{
		KeepAlive:  30,
		CleanStart: true,
	}

	p, err := cemqtt.New(context.TODO(), clientConfig, cp, "test-topic", nil, 0, false)
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}
	defer p.Close(context.Background())

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	for i := 0; i < count; i++ {
		e := cloudevents.NewEvent()
		e.SetID(uuid.New().String())
		e.SetType("com.cloudevents.sample.sent")
		e.SetSource("https://github.com/cloudevents/sdk-go/samples/mqtt/sender")
		_ = e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
			"id":      i,
			"message": "Hello, World!",
		})

		if result := c.Send(
			context.TODO(),
			e,
		); cloudevents.IsUndelivered(result) {
			log.Printf("failed to send: %v", result)
		} else {
			log.Printf("sent: %d, accepted: %t", i, cloudevents.IsACK(result))
		}
		time.Sleep(1 * time.Second)
	}
}
