package nats

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/event"

	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type CodecV02 struct {
	CodecStructured

	Encoding Encoding
}

var _ transport.Codec = (*CodecV02)(nil)

func (v CodecV02) Encode(ctx context.Context, e event.Event) (transport.Message, error) {
	switch v.Encoding {
	case Default:
		fallthrough
	case StructuredV02:
		return v.encodeStructured(ctx, e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", v.Encoding)
	}
}

func (v CodecV02) Decode(ctx context.Context, msg transport.Message) (*event.Event, error) {
	// only structured is supported as of v0.2
	switch v.inspectEncoding(ctx, msg) {
	case StructuredV02:
		return v.decodeStructured(ctx, event.CloudEventsVersionV02, msg)
	default:
		return nil, transport.NewErrMessageEncodingUnknown("v02", TransportName)
	}
}

func (v CodecV02) inspectEncoding(ctx context.Context, msg transport.Message) Encoding {
	version := msg.CloudEventsVersion()
	if version != event.CloudEventsVersionV02 {
		return Unknown
	}
	_, ok := msg.(*Message)
	if !ok {
		return Unknown
	}
	return StructuredV02
}
