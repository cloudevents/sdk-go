package nats

type Encoding int32

const (
	Default Encoding = iota
	StructuredV02
	Unknown
)

func (e Encoding) String() string {
	switch e {
	case Default:
		return "Default Encoding"
	case StructuredV02:
		return "Structured Encoding v0.2"
	default:
		return "Unknown Encoding"
	}
}
