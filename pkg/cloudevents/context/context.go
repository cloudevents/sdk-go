package context

import (
	"context"
	"net/url"
)

// Opaque key type used to store target
type targetKeyType struct{}

var targetKey = targetKeyType{}

func WithTarget(ctx context.Context, target string) context.Context {
	return context.WithValue(ctx, targetKey, target)
}

func TargetFrom(ctx context.Context) *url.URL {
	c := ctx.Value(targetKey)
	if c != nil {
		if target, err := url.Parse(c.(string)); err == nil {
			return target
		}
	}
	return nil
}

// Opaque key type used to store TransportContext
type transportContextKeyType struct{}

var transportContextKey = transportContextKeyType{}

func WithTransportContext(ctx context.Context, tcxt interface{}) context.Context {
	return context.WithValue(ctx, transportContextKey, tcxt)
}

func TransportContextFrom(ctx context.Context) interface{} {
	return ctx.Value(transportContextKey)
}
