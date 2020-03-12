package transport

import (
	"context"
)

// TODO: document
type Engine interface {
	// Blocking call.
	Inbound(ctx context.Context, inbound interface{}) error

	// Blocking call.
	Outbound(ctx context.Context, outbound interface{}) error
}
