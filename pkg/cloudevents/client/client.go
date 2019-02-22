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
	transport transport.Sender
	receiver  Receiver
}

func (c *Client) Send(ctx context.Context, event cloudevents.Event) error {
	if c.transport == nil {
		return fmt.Errorf("client not ready, transport not initalized")
	}
	return c.transport.Send(ctx, event)
}

func (c *Client) Receive(event cloudevents.Event) {
	if c.receiver != nil {
		c.receiver(event)
	}
}

func (c *Client) StartReceiver(ctx context.Context, fn Receiver) error {
	if c.transport == nil {
		return fmt.Errorf("client not ready, transport not initalized")
	}

	if t, ok := c.transport.(*http.Transport); ok {
		return c.startHttpReceiver(ctx, t, fn)
	}

	if t, ok := c.transport.(*nats.Transport); ok {
		return c.startNatsReceiver(ctx, t, fn)
	}

	return fmt.Errorf("unknown transport type: %T", c.transport)
}

func (c *Client) applyClientOptions(opts ...ClientOption) error {
	for _, fn := range opts {
		if err := fn(c); err != nil {
			return err
		}
	}
	return nil
}
