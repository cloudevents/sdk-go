/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"
	"net"

	mqtt_paho "github.com/cloudevents/sdk-go/protocol/mqtt_paho/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/eclipse/paho.golang/paho"
)

func main() {
	ctx := context.Background()
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		log.Fatalf("failed to connect to mqtt broker: %s", err.Error())
	}
	config := &paho.ClientConfig{
		ClientID: "receiver-client-id",
		Conn:     conn,
	}
	subscribeOpt := &paho.Subscribe{
		Subscriptions: map[string]paho.SubscribeOptions{
			"test-topic": {QoS: 0},
		},
	}
	p, err := mqtt_paho.New(ctx, config, mqtt_paho.WithSubscribe(subscribeOpt))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}
	defer p.Close(ctx)

	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("receiver start consuming messages from test-topic\n")
	err = c.StartReceiver(ctx, receive)
	if err != nil {
		log.Fatalf("failed to start receiver: %s", err)
	} else {
		log.Printf("receiver stopped\n")
	}
}

func receive(ctx context.Context, event cloudevents.Event) {
	fmt.Printf("%s", event)
}
