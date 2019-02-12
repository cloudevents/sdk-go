package client

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"net/http"
	"net/url"
)

type Client struct {
	ctx    context.Context
	sender transport.Sender
}

func NewHttpClient(ctx context.Context, targetUrl string, encoding cloudeventshttp.Encoding) (*Client, error) {
	target, err := url.Parse(targetUrl)
	if err != nil {
		return nil, err
	}

	// TODO: context is added to overload the http Method and others. Plumb this.
	req := http.Request{
		Method: http.MethodPost,
		URL:    target,
	}
	ctx = cloudeventshttp.ContextWithValue(ctx, req)

	sender := cloudeventshttp.Transport{Encoding: encoding}

	c := &Client{
		ctx:    ctx,
		sender: &sender,
	}
	return c, nil
}

func (c *Client) Send(event cloudevents.Event) error {
	if c.sender == nil {
		return fmt.Errorf("client not ready, transport not initalized")
	}
	return c.sender.Send(c.ctx, event)
}
