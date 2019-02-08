package http

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type CodecV01 struct {
	Encoding Encoding
}

var _ transport.Codec = (*CodecV01)(nil)

func (v CodecV01) Encode(e canonical.Event) (transport.Message, error) {
	switch v.Encoding {
	case Default:
		fallthrough
	case BinaryV01:
		return v.encodeBinary(e)
	case StructuredV01:
		return v.encodeStructured(e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", v.Encoding)
	}
}

func (v CodecV01) Decode(msg transport.Message) (canonical.Event, error) {
	switch v.inspectEncoding(msg) {
	case BinaryV01:
		return v.decodeBinary(msg)
	case StructuredV01:
		return v.decodeStructured(msg)
	default:
		return nil, fmt.Errorf("unknown encoding for message %v", msg)
	}
}

func (v CodecV01) encodeBinary(e canonical.Event) (transport.Message, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v CodecV01) encodeStructured(e canonical.Event) (transport.Message, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v CodecV01) decodeBinary(msg transport.Message) (canonical.Event, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v CodecV01) decodeStructured(msg transport.Message) (canonical.Event, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v CodecV01) inspectEncoding(msg transport.Message) Encoding {
	return Unknown
}
