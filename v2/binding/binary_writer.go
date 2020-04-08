package binding

import (
	"context"
	"io"
)

// BinaryWriter is used to visit a binary Message and generate a new representation.
//
// Protocols that supports binary encoding should implement this interface to implement direct
// binary to binary encoding and event to binary encoding.
//
// Start() and End() methods are invoked every time this BinaryWriter implementation is used to visit a Message
type BinaryWriter interface {
	MessageMetadataWriter

	// Method invoked at the beginning of the visit. Useful to perform initial memory allocations
	Start(ctx context.Context) error

	// SetData receives an io.Reader for the data attribute.
	// io.Reader is not invoked when the data attribute is empty
	SetData(data io.Reader) error

	// End method is invoked only after the whole encoding process ends successfully.
	// If it fails, it's never invoked. It can be used to finalize the message.
	End(ctx context.Context) error
}
