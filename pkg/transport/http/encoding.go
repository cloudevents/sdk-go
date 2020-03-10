package http

// Encoding to use for HTTP transport.
type Encoding int32

const (
	// Default is binary
	Default Encoding = iota

	// Binary
	Binary

	// Structured
	Structured
)
