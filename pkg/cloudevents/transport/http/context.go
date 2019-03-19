package http

import (
	"context"
	"strings"
)

// TransportContext allows a Receiver to understand the context of a request.
type TransportContext struct {
	URI    string
	Host   string
	Method string
}

func (tx TransportContext) String() string {
	b := strings.Builder{}

	b.WriteString("Transport Context,\n")
	b.WriteString("  URI: " + tx.URI + "\n")
	b.WriteString("  Host: " + tx.Host + "\n")
	b.WriteString("  Method: " + tx.Method + "\n")

	return b.String()
}

// Opaque key type used to store TransportContext
type transportContextKeyType struct{}

var transportContextKey = transportContextKeyType{}

func WithTransportContext(ctx context.Context, tcxt TransportContext) context.Context {
	return context.WithValue(ctx, transportContextKey, tcxt)
}

func TransportContextFrom(ctx context.Context) TransportContext {
	tctx := ctx.Value(transportContextKey)
	if tctx != nil {
		if tx, ok := tctx.(TransportContext); ok {
			return tx
		}
		if tx, ok := tctx.(*TransportContext); ok {
			return *tx
		}
	}
	return TransportContext{}
}
