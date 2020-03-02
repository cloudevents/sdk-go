package nats

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/event"

	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type CodecV03 struct {
	CodecStructured

	Encoding Encoding
}

var _ transport.Codec = (*CodecV03)(nil)

func (v CodecV03) Encode(ctx context.Context, e event.Event) (transport.Message, error) {
	switch v.Encoding {
	case Default:
		fallthrough
	case StructuredV03:
		return v.encodeStructured(ctx, e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", v.Encoding)
	}
}

func (v CodecV03) Decode(ctx context.Context, msg transport.Message) (*event.Event, error) {
	// only structured is supported as of v0.3
	switch v.inspectEncoding(ctx, msg) {
	case StructuredV03:
		return v.decodeStructured(ctx, event.CloudEventsVersionV03, msg)
	default:
		return nil, transport.NewErrMessageEncodingUnknown("v03", TransportName)
	}
}

func (v CodecV03) inspectEncoding(ctx context.Context, msg transport.Message) Encoding {
	version := msg.CloudEventsVersion()
	if version != event.CloudEventsVersionV03 {
		return Unknown
	}
	_, ok := msg.(*Message)
	if !ok {
		return Unknown
	}
	return StructuredV03
}
