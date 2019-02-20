package client

import (
	"context"
	"fmt"
	cloudeventsnats "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
	"github.com/nats-io/go-nats"
	"log"
)

func NewNatsClient(ctx context.Context, natsServer, subject string, encoding cloudeventsnats.Encoding) (*Client, error) {
	// TODO: context is added to overload defaults. Plumb this.
	conn, err := nats.Connect(natsServer)
	if err != nil {
		return nil, err
	}
	transport := cloudeventsnats.Transport{
		Conn:     conn,
		Encoding: encoding,
	}
	// add subject
	ctx = cloudeventsnats.ContextWithSubject(ctx, subject)
	c := &Client{
		ctx:       ctx,
		transport: &transport,
	}
	return c, nil
}

func (c *Client) startNatsReceiver(t *cloudeventsnats.Transport, fn Receiver) error {
	if t.Conn == nil {
		return fmt.Errorf("nats connection is required to be set")
	}
	if c.receiver != nil {
		return fmt.Errorf("client already has a receiver")
	}
	if t.Receiver != nil {
		return fmt.Errorf("transport already has a receiver")
	}

	c.receiver = fn
	t.Receiver = c

	subject := cloudeventsnats.SubjectFromContext(c.ctx)
	if subject == "" {
		return fmt.Errorf("subject is required for nats receiver")
	}

	go func() {
		if err := t.Listen(c.ctx, subject); err != nil {
			log.Fatalf("failed to listen, %s", err.Error())
		}
	}()

	return nil
}
