package main

import (
	"context"
	"log"
	"net/http"

	cews "github.com/cloudevents/sdk-go/protocol/ws/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
)

func main() {
	ctx := context.Background()

	err := http.ListenAndServe(":8080", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		p, err := cews.Accept(ctx, writer, request, nil)
		if err != nil {
			log.Printf("Error while accepting a websocket request: %v\n", err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		defer p.Close(ctx)

		message, err := p.Receive(ctx)
		if err != nil {
			log.Printf("Error while receiving a websocket message: %v\n", err)
			return
		}

		receivedEvent, err := binding.ToEvent(ctx, message)
		if err != nil {
			log.Printf("Error while parsing the message: %v\n", err)
			return
		}

		log.Printf("Received event:\n%s", receivedEvent)

		// Echo the event back
		err = p.Send(ctx, binding.ToMessage(receivedEvent))
		if err != nil {
			log.Printf("Error while echoing the event back: %v\n", err)
			return
		}
	}))
	if err != nil {
		log.Fatalf("failed to start the listener: %v\n", err)
	}
}
