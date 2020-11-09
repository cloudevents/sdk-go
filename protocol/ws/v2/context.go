package v2

import (
	"context"

	"nhooyr.io/websocket"
)

type codeKey struct{}

type reasonKey struct{}

func WithCloseReason(ctx context.Context, code websocket.StatusCode, reason string) context.Context {
	return context.WithValue(context.WithValue(ctx, codeKey{}, code), reasonKey{}, reason)
}
