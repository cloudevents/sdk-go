package transport

import (
	"context"
)

// Closer is the common interface for things that can be closed
type Closer interface {
	Close(ctx context.Context) error
}
