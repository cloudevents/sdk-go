package transport

import (
	"context"
)

// TODO: document
type Engine interface {
	// Blocking call. Context is used to cancel.
	StartInbound(ctx context.Context) error
}
