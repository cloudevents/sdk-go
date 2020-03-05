package binding

import "context"

// Receiver receives messages.
type Receiver interface {
	// Receive blocks till a message is received or ctx expires.
	//
	// A non-nil error means the receiver is closed.
	// io.EOF means it closed cleanly, any other value indicates an error.
	Receive(ctx context.Context) (Message, error)
}

// ReceiveCloser is a Receiver that can be closed.
type ReceiveCloser interface {
	Receiver
	Closer
}
