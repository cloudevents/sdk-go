package http

// Encoding to use for HTTP transport.
type Encoding int32

// TODO: there is some legacy stuff in here still, clean it up.

const (
	// Default is binary
	Default Encoding = iota

	// Binary
	Binary

	// Structured
	Structured
)
