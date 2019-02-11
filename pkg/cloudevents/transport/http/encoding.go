package http

type Encoding int32

const (
	Default Encoding = iota
	BinaryV01
	StructuredV01
	BinaryV02
	StructuredV02
	Unknown
)

const (
	CloudEventsStructuredMediaType string = "application/cloudevents+json"
)

func (e Encoding) String() string {
	switch e {
	case Default:
		return "Default Encoding"
	case BinaryV01:
		return "Binary Encoding v0.1"
	case StructuredV01:
		return "Structured Encoding v0.1"
	case BinaryV02:
		return "Binary Encoding v0.2"
	case StructuredV02:
		return "Structured Encoding v0.2"
	default:
		return "Unknown Encoding"
	}
}
