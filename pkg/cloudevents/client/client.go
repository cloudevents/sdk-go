package client

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// Receive is the signature of a fn to be invoked for incoming cloudevents.
// If fn returns an error, EventResponse will not be considered by the client or
// or transport.
type Receive func(ctx context.Context, event cloudevents.Event, resp *cloudevents.EventResponse) error

type Client interface {
	Send(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error)

	StartReceiver(ctx context.Context, fn Receive) error
	StopReceiver(ctx context.Context) error
}

type ceClient struct {
	transport transport.Transport
	receive   Receive

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

// Receive is called from from the transport on event delivery.
func (c *ceClient) Receive(ctx context.Context, event cloudevents.Event, resp *cloudevents.EventResponse) error {
	if c.receive != nil {
		err := c.receive(ctx, event, resp)
		// Apply the defaulter chain to the outgoing event.
		if err == nil && resp != nil && resp.Event != nil && len(c.eventDefaulterFns) > 0 {
			for _, fn := range c.eventDefaulterFns {
				*resp.Event = fn(*resp.Event)
			}
			// Validate the event conforms to the CloudEvents Spec.
			if err := resp.Event.Validate(); err != nil {
				return fmt.Errorf("cloudevent validation failed on response event: %v", err)
			}
		}
		return err
	}
	return nil
}

func (c *ceClient) StartReceiver(ctx context.Context, fn Receive) error {
	if c.transport == nil {
		return fmt.Errorf("client not ready, transport not initialized")
	}
	if c.receive != nil {
		return fmt.Errorf("client already has a receiver")
	}

	c.receive = fn

	return c.transport.StartReceiver(ctx)
}

func (c *ceClient) StopReceiver(ctx context.Context) error {
	if c.transport == nil {
		return fmt.Errorf("client not ready, transport not initialized")
	}

	err := c.transport.StopReceiver(ctx)
	c.receive = nil
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
