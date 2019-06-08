package stan

import (
	"fmt"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/codec"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type CodecV02 struct {
	Encoding Encoding
}

var _ transport.Codec = (*CodecV02)(nil)

func (v CodecV02) Encode(e cloudevents.Event) (transport.Message, error) {
	switch v.Encoding {
	case Default:
		fallthrough
	case StructuredV02:
		return v.encodeStructured(e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", v.Encoding)
	}
}

func (v CodecV02) Decode(msg transport.Message) (*cloudevents.Event, error) {
	// only structured is supported as of v0.2
	switch v.inspectEncoding(msg) {
	case StructuredV02:
		return v.decodeStructured(msg)
	default:
		return nil, transport.NewErrMessageEncodingUnknown("v02", TransportName)
	}
}

func (v CodecV02) encodeStructured(e cloudevents.Event) (transport.Message, error) {
	body, err := codec.JsonEncodeV02(e)
	if err != nil {
		return nil, err
	}
	return &Message{
		Body: body,
	}, nil
}

func (v CodecV02) decodeStructured(msg transport.Message) (*cloudevents.Event, error) {
	m, ok := msg.(*Message)
	if !ok {
		return nil, fmt.Errorf("failed to convert transport.Message to nats.Message")
	}
	return codec.JsonDecodeV02(m.Body)
}

func (v CodecV02) inspectEncoding(msg transport.Message) Encoding {
	version := msg.CloudEventsVersion()
	if version != cloudevents.CloudEventsVersionV02 {
		return Unknown
	}
	_, ok := msg.(*Message)
	if !ok {
		return Unknown
	}
	return StructuredV02
}
