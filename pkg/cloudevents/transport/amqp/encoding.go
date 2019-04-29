package amqp

// Encoding to use for amqp transport.
type Encoding int32

const (
	// Default allows amqp transport implementation to pick.
	Default Encoding = iota
	// BinaryV02 is Binary CloudEvents spec v0.2.
	BinaryV02
	// StructuredV02 is Structured CloudEvents spec v0.2.
	StructuredV02
	// BinaryV03 is Binary CloudEvents spec v0.3.
	BinaryV03
	// StructuredV03 is Structured CloudEvents spec v0.3.
	StructuredV03
	// Unknown is unknown.
	Unknown
)

// String pretty-prints the encoding as a string.
func (e Encoding) String() string {
	switch e {
	case Default:
		return "Default Encoding " + e.Version()

	// Binary
	case BinaryV02:
		fallthrough
	case BinaryV03:
		return "Binary Encoding " + e.Version()

	// Structured
	case StructuredV02:
		fallthrough
	case StructuredV03:
		return "Structured Encoding " + e.Version()

	default:
		return "Unknown Encoding"
	}
}

// Version pretty-prints the encoding version as a string.
func (e Encoding) Version() string {
	switch e {

	// Version 0.2
	case Default: // <-- Move when a new default is wanted.
		fallthrough
	case StructuredV02:
		return "v0.2"

	// Version 0.3
	case StructuredV03:
		return "v0.3"

	// Unknown
	default:
		return "Unknown"
	}
}
