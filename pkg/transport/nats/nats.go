package nats

// Encoding to use for NATS transport.
type Encoding int32

const (
	// Default allows NATS transport implementation to pick.
	Default Encoding = iota
	// StructuredV03 is Structured CloudEvents spec v0.3.
	StructuredV03
	// StructuredV1 is Structured CloudEvents spec v1.0.
	StructuredV1
	// Unknown is unknown.
	Unknown
)

// String pretty-prints the encoding as a string.
func (e Encoding) String() string {
	switch e {
	case Default:
		return "Default Encoding " + e.Version()

	// Structured
	case StructuredV03, StructuredV1:
		return "Structured Encoding " + e.Version()

	default:
		return "Unknown Encoding"
	}
}

// Version pretty-prints the encoding version as a string.
func (e Encoding) Version() string {
	switch e {
	// Version 0.3
	case StructuredV03:
		return "v0.3"

	// Version 1.0
	case StructuredV1, Default:
		return "v1.0"

	// Unknown
	default:
		return "Unknown"
	}
}
