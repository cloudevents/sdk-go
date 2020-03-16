package http

// Encoding to use for HTTP transport.
type Encoding int32

const (
	// Default is binary
	// Deprecated: use binding
	Default Encoding = iota

	// Binary
	// Deprecated: use binding
	Binary

	// Structured
	// Deprecated: use binding
	Structured
)
