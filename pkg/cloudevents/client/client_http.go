package client

import (
	"context"
	"fmt"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"log"
	"net/http"
	"net/url"
)

func NewHttpClient(ctx context.Context, targetUrl string, encoding cloudeventshttp.Encoding) (*Client, error) {
	var target *url.URL
	if targetUrl != "" {
		var err error
		target, err = url.Parse(targetUrl)
		if err != nil {
			return nil, err
		}
	}

	// TODO: context is added to overload the http Method and others. Plumb this.
	req := http.Request{
		Method: http.MethodPost,
		URL:    target,
	}
	ctx = cloudeventshttp.ContextWithRequest(ctx, req)

	transport := cloudeventshttp.Transport{Encoding: encoding}

	c := &Client{
		ctx:       ctx,
		transport: &transport,
	}
	return c, nil
}

func StartHttpReceiver(ctx *context.Context, fn Receiver) error {
	if ctx == nil {
		return fmt.Errorf("context object required for Client.StartReceiver")
	}
	c, err := NewHttpClient(*ctx, "", 0)
	if err != nil {
		return err
	}
	*ctx = ContextWithClient(*ctx, c)

	if err := c.StartReceiver(fn); err != nil {
		return err
	}
	return nil
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
	port := PortFromContext(c.ctx)

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), t))
	}()

	return nil
}
