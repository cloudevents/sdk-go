package client

import (
	"context"
	"fmt"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"log"
	"net/http"
)

func NewHttpClient(opts ...ClientOption) (*Client, error) {
	transport := cloudeventshttp.Transport{}

	c := &Client{
		ctx:       context.Background(),
		transport: &transport,
	}

	// Default request method.
	req := http.Request{
		Method: http.MethodPost,
	}
	c.ctx = cloudeventshttp.ContextWithRequest(c.ctx, req)

	if err := c.applyClientOptions(opts...); err != nil {
		return nil, err
	}

	return c, nil
}

func StartHttpReceiver(fn Receiver, opts ...ClientOption) (context.Context, error) {
	c, err := NewHttpClient(opts...)
	if err != nil {
		return nil, err
	}
	ctx := ContextWithClient(c.ctx, c)

	if err := c.StartReceiver(fn); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (c *Client) startHttpReceiver(t *cloudeventshttp.Transport, fn Receiver) error {
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
