package client

import (
	"context"
	"fmt"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"log"
	"net/http"
)

func NewHTTPClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		transport: &cloudeventshttp.Transport{
			// Default the request method.
			Req: &http.Request{
				Method: http.MethodPost,
			},
		},
	}

	if err := c.applyClientOptions(opts...); err != nil {
		return nil, err
	}

	return c, nil
}

func StartHTTPReceiver(ctx context.Context, fn Receiver, opts ...ClientOption) (*Client, error) {
	c, err := NewHTTPClient(opts...)
	if err != nil {
		return nil, err
	}

	if err := c.StartReceiver(ctx, fn); err != nil {
		return c, err
	}
	return c, nil
}

func (c *Client) startHTTPReceiver(ctx context.Context, t *cloudeventshttp.Transport, fn Receiver) error {
	if c.receiver != nil {
		return fmt.Errorf("client already has a receiver")
	}
	if t.Receiver != nil {
		return fmt.Errorf("transport already has a receiver")
	}
	c.receiver = fn
	t.Receiver = c

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", t.GetPort()), t))
	}()

	return nil
}
