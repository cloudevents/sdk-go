package client

import (
	"context"
)

// Opaque key type used to store Client Pointer
type clientKeyType struct{}

var clientKey = clientKeyType{}

func ContextWithClient(ctx context.Context, c *Client) context.Context {
	return context.WithValue(ctx, clientKey, c)
}

func ClientFromContext(ctx context.Context) *Client {
	c := ctx.Value(clientKey)
	if c != nil {
		return c.(*Client)
	}
	return nil
}
