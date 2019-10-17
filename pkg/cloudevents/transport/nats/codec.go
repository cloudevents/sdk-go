package nats

import (
	"context"
	"fmt"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type Codec struct {
	Encoding Encoding

	v02 *CodecV02
	v03 *CodecV03
	v1  *CodecV1
}

var _ transport.Codec = (*Codec)(nil)

func (c *Codec) Encode(ctx context.Context, e cloudevents.Event) (transport.Message, error) {
	switch c.Encoding {
	case Default:
		fallthrough
	case StructuredV02:
		if c.v02 == nil {
			c.v02 = &CodecV02{Encoding: c.Encoding}
		}
		return c.v02.Encode(ctx, e)
	case StructuredV03:
		if c.v03 == nil {
			c.v03 = &CodecV03{Encoding: c.Encoding}
		}
		return c.v03.Encode(ctx, e)
	case StructuredV1:
		if c.v1 == nil {
			c.v1 = &CodecV1{Encoding: c.Encoding}
		}
		return c.v1.Encode(ctx, e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", c.Encoding)
	}
}

func (c *Codec) Decode(ctx context.Context, msg transport.Message) (*cloudevents.Event, error) {
	switch c.inspectEncoding(ctx, msg) {
	case Default:
		fallthrough
	case StructuredV02:
		if c.v02 == nil {
			c.v02 = &CodecV02{Encoding: c.Encoding}
		}
		return c.v02.Decode(ctx, msg)
	case StructuredV03:
		if c.v03 == nil {
			c.v03 = &CodecV03{Encoding: c.Encoding}
		}
		return c.v03.Decode(ctx, msg)
	case StructuredV1:
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

	if c.v02 == nil {
		c.v02 = &CodecV02{Encoding: c.Encoding}
	}
	// Try v0.2.
	encoding = c.v02.inspectEncoding(ctx, msg)
	if encoding != Unknown {
		return encoding
	}

	// We do not understand the message encoding.
	return Unknown
}
