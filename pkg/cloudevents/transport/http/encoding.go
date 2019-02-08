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
