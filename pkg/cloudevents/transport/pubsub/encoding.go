package pubsub

import (
	"context"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
)

// Encoding to use for pubsub transport.
type Encoding int32

type EncodingSelector func(context.Context, cloudevents.Event) Encoding

const (
	// Default allows pubsub transport implementation to pick.
	Default Encoding = iota
	// BinaryV03 is Binary CloudEvents spec v0.3.
	BinaryV03
	// BinaryV1 is Binary CloudEvents spec v1.0.
	BinaryV1
	// StructuredV03 is Structured CloudEvents spec v0.3.
	StructuredV03
	// StructuredV1 is Structured CloudEvents spec v1.0.
	StructuredV1
	// Unknown is unknown.
	Unknown
)

// DefaultBinaryEncodingSelectionStrategy implements a selection process for
// which binary encoding to use based on spec version of the event.
func DefaultBinaryEncodingSelectionStrategy(ctx context.Context, e cloudevents.Event) Encoding {
	switch e.SpecVersion() {
	case cloudevents.CloudEventsVersionV01, cloudevents.CloudEventsVersionV02, cloudevents.CloudEventsVersionV03:
		return BinaryV03
	case cloudevents.CloudEventsVersionV1:
		return BinaryV1
	}
	// Unknown version, return Default.
	return Default
}

// DefaultStructuredEncodingSelectionStrategy implements a selection process
// for which structured encoding to use based on spec version of the event.
func DefaultStructuredEncodingSelectionStrategy(ctx context.Context, e cloudevents.Event) Encoding {
	switch e.SpecVersion() {
	case cloudevents.CloudEventsVersionV01, cloudevents.CloudEventsVersionV02, cloudevents.CloudEventsVersionV03:
		return StructuredV03
	case cloudevents.CloudEventsVersionV1:
		return StructuredV1
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

	// Version 0.2
	// Version 0.3
	case Default, BinaryV03, StructuredV03:
		return "v0.3"

		// Version 1.0
	case BinaryV1, StructuredV1:
		return "v1.0"

	// Unknown
	default:
		return "Unknown"
	}
}
