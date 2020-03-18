package nats

import (
	"context"
	"fmt"

	"github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/transport"
)

type CodecV03 struct {
	CodecStructured

	Encoding Encoding
}

var _ transport.Codec = (*CodecV03)(nil)

func (v CodecV03) Encode(ctx context.Context, e cloudevents.Event) (transport.Message, error) {
	switch v.Encoding {
	case Default:
		fallthrough
	case StructuredV03:
		return v.encodeStructured(ctx, e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", v.Encoding)
	}
}

func (v CodecV03) Decode(ctx context.Context, msg transport.Message) (*cloudevents.Event, error) {
	// only structured is supported as of v0.3
	switch v.inspectEncoding(ctx, msg) {
	case StructuredV03:
		return v.decodeStructured(ctx, cloudevents.CloudEventsVersionV03, msg)
	default:
		return nil, transport.NewErrMessageEncodingUnknown("v03", TransportName)
	}
}

func (v CodecV03) inspectEncoding(ctx context.Context, msg transport.Message) Encoding {
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
