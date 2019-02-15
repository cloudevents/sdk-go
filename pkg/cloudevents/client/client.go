package client

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	cloudeventsnats "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
	"github.com/nats-io/go-nats"
	"log"
	"net/http"
	"net/url"
)

type Receiver func(event cloudevents.Event)

type Client struct {
	ctx      context.Context
	sender   transport.Sender
	receiver Receiver
}

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

func (c *Client) Receive(event cloudevents.Event) {
	if c.receiver != nil {
		c.receiver(event)
	}
}

// Opaque key type used to store Http Request
type portKeyType struct{}

var portKey = portKeyType{}

func ContextWithPort(ctx context.Context, port int) context.Context {
	return context.WithValue(ctx, portKey, port)
}

func PortFromContext(ctx context.Context) int {
	port := ctx.Value(portKey)
	if port != nil {
		return port.(int)
	}
	return 8080 // default
}

func (c *Client) StartReceiver(fn Receiver) error {
	if c.sender == nil {
		return fmt.Errorf("client not ready, transport not initalized")
	}

	if t, ok := c.sender.(*cloudeventshttp.Transport); ok {
		if t.Receiver != nil {
			return fmt.Errorf("client already has a receiver")
		}
		c.receiver = fn
		t.Receiver = c
		port := PortFromContext(c.ctx)
		go log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), t))
	}

	// TODO: nats
	return nil
}
