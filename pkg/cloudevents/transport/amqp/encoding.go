package amqp

// Encoding to use for amqp transport.
type Encoding int32

const (
	// Default allows amqp transport implementation to pick.
	Default Encoding = iota
	// BinaryV03 is Binary CloudEvents spec v0.3.
	BinaryV03
	// StructuredV03 is Structured CloudEvents spec v0.3.
	StructuredV03
	// BinaryV1 is Binary CloudEvents spec v1.0.
	BinaryV1
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

	// Binary
	case BinaryV03, BinaryV1:
		return "Binary Encoding " + e.Version()

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
