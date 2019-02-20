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
