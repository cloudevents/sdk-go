package pubsub

import "github.com/cloudevents/sdk-go/pkg/cloudevents"

// Encoding to use for pubsub transport.
type Encoding int32

type EncodingSelector func(e cloudevents.Event) Encoding

const (
	// Default allows pubsub transport implementation to pick.
	Default Encoding = iota
	// BinaryV03 is Binary CloudEvents spec v0.3.
	BinaryV03
	// StructuredV03 is Structured CloudEvents spec v0.3.
	StructuredV03
	// Unknown is unknown.
	Unknown
)

// DefaultBinaryEncodingSelectionStrategy implements a selection process for
// which binary encoding to use based on spec version of the event.
func DefaultBinaryEncodingSelectionStrategy(e cloudevents.Event) Encoding {
	switch e.SpecVersion() {
	case cloudevents.CloudEventsVersionV01, cloudevents.CloudEventsVersionV02, cloudevents.CloudEventsVersionV03:
		return BinaryV03
	}
	// Unknown version, return Default.
	return Default
}

// DefaultStructuredEncodingSelectionStrategy implements a selection process
// for which structured encoding to use based on spec version of the event.
func DefaultStructuredEncodingSelectionStrategy(e cloudevents.Event) Encoding {
	switch e.SpecVersion() {
	case cloudevents.CloudEventsVersionV01, cloudevents.CloudEventsVersionV02, cloudevents.CloudEventsVersionV03:
		return StructuredV03
	}
	// Unknown version, return Default.
	return Default
}

// String pretty-prints the encoding as a string.
func (e Encoding) String() string {
	switch e {
	case Default:
		return "Default Encoding " + e.Version()

	// Binary
	case BinaryV03:
		return "Binary Encoding " + e.Version()

	// Structured
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

	// Version 0.3
	case StructuredV03:
		return "v0.3"

	// Unknown
	default:
		return "Unknown"
	}
}
