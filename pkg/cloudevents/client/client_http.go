package client

import (
	"context"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"net/http"
)

func NewHTTPClient(opts ...Option) (Client, error) {
	c := &ceClient{}
	c.transport = &cloudeventshttp.Transport{
		// Default the request method.
		Req: &http.Request{
			Method: http.MethodPost,
		},
		Receiver: c,
	}

	if err := c.applyOptions(opts...); err != nil {
		return nil, err
	}
	return c, nil
}

func StartHTTPReceiver(ctx context.Context, fn Receiver, opts ...Option) (Client, error) {
	c, err := NewHTTPClient(opts...)
	if err != nil {
		return nil, err
	}

	if err := c.StartReceiver(ctx, fn); err != nil {
		return nil, err
	}
	return c, nil
}
