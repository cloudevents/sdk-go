package client

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type Receiver func(event cloudevents.Event)

type Client interface {
	Send(ctx context.Context, event cloudevents.Event) error

	StartReceiver(ctx context.Context, fn Receiver) (context.Context, error)
	StopReceiver(ctx context.Context) error

	Receive(event cloudevents.Event)
}

type ceClient struct {
	transport transport.Sender
	receiver  Receiver

	eventDefaulterFns []EventDefaulter
}

func (c *ceClient) Send(ctx context.Context, event cloudevents.Event) error {
	// Confirm we have a transport set.
	if c.transport == nil {
		return fmt.Errorf("client not ready, transport not initialized")
	}
	// Apply the defaulter chain to the incoming event.
	if len(c.eventDefaulterFns) > 0 {
		for _, fn := range c.eventDefaulterFns {
			event = fn(event)
		}
	}
	// Validate the event conforms to the CloudEvents Spec.
	if err := event.Validate(); err != nil {
		return err
	}
	// Send the event over the transport.
	return c.transport.Send(ctx, event)
}

func (c *ceClient) Receive(event cloudevents.Event) {
	if c.receiver != nil {
		c.receiver(event)
	}
}

func (c *ceClient) StartReceiver(ctx context.Context, fn Receiver) (context.Context, error) {
	if c.transport == nil {
		return ctx, fmt.Errorf("client not ready, transport not initialized")
	}
	if c.receiver != nil {
		return ctx, fmt.Errorf("client already has a receiver")
	}

	c.receiver = fn

	return c.transport.StartReceiver(ctx)
}

func (c *ceClient) StopReceiver(ctx context.Context) error {
	if c.transport == nil {
		return fmt.Errorf("client not ready, transport not initialized")
	}

	err := c.transport.StopReceiver(ctx)
	c.receiver = nil
	return err
}

func (c *ceClient) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(c); err != nil {
			return err
		}
	}
	return nil
}
