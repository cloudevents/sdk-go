package client

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	cloudeventsnats "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
	"github.com/nats-io/go-nats"
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
	ctx = cloudeventshttp.ContextWithRequest(ctx, req)

	sender := cloudeventshttp.Transport{Encoding: encoding}

	c := &Client{
		ctx:    ctx,
		sender: &sender,
	}
	return c, nil
}

func NewNatsClient(ctx context.Context, natsServer, subject string) (*Client, error) {
	// TODO: context is added to overload defaults. Plumb this.
	conn, err := nats.Connect(natsServer)
	if err != nil {
		return nil, err
	}
	sender := cloudeventsnats.Transport{
		Conn: conn,
	}
	// add subject
	ctx = cloudeventsnats.ContextWithSubject(ctx, subject)
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
