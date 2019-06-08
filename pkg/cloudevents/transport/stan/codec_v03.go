package stan

import (
	"fmt"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/codec"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type CodecV03 struct {
	Encoding Encoding
}

var _ transport.Codec = (*CodecV03)(nil)

func (v CodecV03) Encode(e cloudevents.Event) (transport.Message, error) {
	switch v.Encoding {
	case Default:
		fallthrough
	case StructuredV03:
		return v.encodeStructured(e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", v.Encoding)
	}
}

func (v CodecV03) Decode(msg transport.Message) (*cloudevents.Event, error) {
	// only structured is supported as of v0.3
	switch v.inspectEncoding(msg) {
	case StructuredV03:
		return v.decodeStructured(msg)
	default:
		return nil, transport.NewErrMessageEncodingUnknown("v03", TransportName)
	}
}

func (v CodecV03) encodeStructured(e cloudevents.Event) (transport.Message, error) {
	body, err := codec.JsonEncodeV03(e)
	if err != nil {
		return nil, err
	}
	return &Message{
		Body: body,
	}, nil
}

func (v CodecV03) decodeStructured(msg transport.Message) (*cloudevents.Event, error) {
	m, ok := msg.(*Message)
	if !ok {
		return nil, fmt.Errorf("failed to convert transport.Message to nats.Message")
	}
	return codec.JsonDecodeV03(m.Body)
}

func (v CodecV03) inspectEncoding(msg transport.Message) Encoding {
	version := msg.CloudEventsVersion()
	if version != cloudevents.CloudEventsVersionV03 {
		return Unknown
	}
	_, ok := msg.(*Message)
	if !ok {
		return Unknown
	}
	return StructuredV03
}
