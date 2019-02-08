package http

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type CodecV02 struct {
	Encoding Encoding
}

var _ transport.Codec = (*CodecV02)(nil)

func (v CodecV02) Encode(e canonical.Event) (transport.Message, error) {
	switch v.Encoding {
	case Default:
		fallthrough
	case BinaryV02:
		return v.encodeBinary(e)
	case StructuredV02:
		return v.encodeStructured(e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", v.Encoding)
	}
}

func (v CodecV02) Decode(msg transport.Message) (*canonical.Event, error) {
	switch v.inspectEncoding(msg) {
	case BinaryV02:
		return v.decodeBinary(msg)
	case StructuredV02:
		return v.decodeStructured(msg)
	default:
		return nil, fmt.Errorf("unknown encoding for message %v", msg)
	}
}

func (v CodecV02) encodeBinary(e canonical.Event) (transport.Message, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v CodecV02) encodeStructured(e canonical.Event) (transport.Message, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v CodecV02) decodeBinary(msg transport.Message) (*canonical.Event, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v CodecV02) decodeStructured(msg transport.Message) (*canonical.Event, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v CodecV02) inspectEncoding(msg transport.Message) Encoding {
	return Unknown
}
