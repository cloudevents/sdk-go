package nats

import (
	"context"
	"fmt"

	"github.com/cloudevents/sdk-go/pkg/event"

	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type Codec struct {
	Encoding Encoding

	v03 *CodecV03
	v1  *CodecV1
}

var _ transport.Codec = (*Codec)(nil)

func (c *Codec) Encode(ctx context.Context, e event.Event) (transport.Message, error) {
	switch c.Encoding {
	case StructuredV03:
		if c.v03 == nil {
			c.v03 = &CodecV03{Encoding: c.Encoding}
		}
		return c.v03.Encode(ctx, e)
	case StructuredV1, Default:
		if c.v1 == nil {
			c.v1 = &CodecV1{Encoding: c.Encoding}
		}
		return c.v1.Encode(ctx, e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", c.Encoding)
	}
}

func (c *Codec) Decode(ctx context.Context, msg transport.Message) (*event.Event, error) {
	switch c.inspectEncoding(ctx, msg) {
	case StructuredV03:
		if c.v03 == nil {
			c.v03 = &CodecV03{Encoding: c.Encoding}
		}
		return c.v03.Decode(ctx, msg)
	case StructuredV1, Default:
		if c.v1 == nil {
			c.v1 = &CodecV1{Encoding: c.Encoding}
		}
		return c.v1.Decode(ctx, msg)
	default:
		return nil, transport.NewErrMessageEncodingUnknown("wrapper", TransportName)
	}
}

func (c *Codec) inspectEncoding(ctx context.Context, msg transport.Message) Encoding {

	if c.v1 == nil {
		c.v1 = &CodecV1{Encoding: c.Encoding}
	}
	// Try v1.0.
	encoding := c.v1.inspectEncoding(ctx, msg)
	if encoding != Unknown {
		return encoding
	}

	if c.v03 == nil {
		c.v03 = &CodecV03{Encoding: c.Encoding}
	}
	// Try v0.3.
	encoding = c.v03.inspectEncoding(ctx, msg)
	if encoding != Unknown {
		return encoding
	}

	// We do not understand the message encoding.
	return Unknown
}
