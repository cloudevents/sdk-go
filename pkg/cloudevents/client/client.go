package client

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
)

type Receiver func(event cloudevents.Event)

type Client struct {
	ctx       context.Context
	transport transport.Sender
	receiver  Receiver
}

func (c *Client) Send(event cloudevents.Event) error {
	if c.transport == nil {
		return fmt.Errorf("client not ready, transport not initalized")
	}
	return c.transport.Send(c.ctx, event)
}

func (c *Client) Receive(event cloudevents.Event) {
	if c.receiver != nil {
		c.receiver(event)
	}
}

func (c *Client) StartReceiver(fn Receiver) error {
	if c.transport == nil {
		return fmt.Errorf("client not ready, transport not initalized")
	}

	if t, ok := c.transport.(*http.Transport); ok {
		return c.startHttpReceiver(t, fn)
	}

	if t, ok := c.transport.(*nats.Transport); ok {
		return c.startNatsReceiver(t, fn)
	}

	return fmt.Errorf("unknown transport type: %T", c.transport)
}
