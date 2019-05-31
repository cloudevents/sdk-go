package pubsub

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/codec"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

const (
	StructuredContentType = "Content-Type"
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
	data, err := codec.JsonEncodeV03(e)
	if err != nil {
		return nil, err
	}

	return &Message{
		Attributes: map[string]string{StructuredContentType: cloudevents.ApplicationCloudEventsJSON},
		Data:       data,
	}, nil
}

func (v CodecV03) decodeStructured(msg transport.Message) (*cloudevents.Event, error) {
	m, ok := msg.(*Message)
	if !ok {
		return nil, fmt.Errorf("failed to convert transport.Message to pubsub.Message")
	}
	return codec.JsonDecodeV03(m.Data)
}

func (v CodecV03) inspectEncoding(msg transport.Message) Encoding {
	version := msg.CloudEventsVersion()
	if version != cloudevents.CloudEventsVersionV03 {
		return Unknown
	}
	m, ok := msg.(*Message)
	if !ok {
		return Unknown
	}
	if m.Attributes[StructuredContentType] == cloudevents.ApplicationCloudEventsJSON {
		return StructuredV03
	}
	return BinaryV03
}
