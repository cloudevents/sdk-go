package client

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type Receiver func(event cloudevents.Event) (*cloudevents.Event, error)

type Client interface {
	Send(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error)

	StartReceiver(ctx context.Context, fn Receiver) error
	StopReceiver(ctx context.Context) error
}

type ceClient struct {
	transport transport.Transport
	receiver  Receiver

	eventDefaulterFns []EventDefaulter
}

func New(t transport.Transport, opts ...Option) (Client, error) {
	c := &ceClient{
		transport: t,
	}
	if err := c.applyOptions(opts...); err != nil {
		return nil, err
	}
	t.SetReceiver(c)
	return c, nil
}

func (c *ceClient) Send(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	// Confirm we have a transport set.
	if c.transport == nil {
		return nil, fmt.Errorf("client not ready, transport not initialized")
	}
	// Apply the defaulter chain to the incoming event.
	if len(c.eventDefaulterFns) > 0 {
		for _, fn := range c.eventDefaulterFns {
			event = fn(event)
		}
	}
	// Validate the event conforms to the CloudEvents Spec.
	if err := event.Validate(); err != nil {
		return nil, err
	}
	// Send the event over the transport.
	return c.transport.Send(ctx, event)
}

func (c *ceClient) Receive(event cloudevents.Event) (*cloudevents.Event, error) {
	if c.receiver != nil {
		return c.receiver(event)
	}
	return nil, nil
}

func (c *ceClient) StartReceiver(ctx context.Context, fn Receiver) error {
	if c.transport == nil {
		return fmt.Errorf("client not ready, transport not initialized")
	}
	if c.receiver != nil {
		return fmt.Errorf("client already has a receiver")
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
