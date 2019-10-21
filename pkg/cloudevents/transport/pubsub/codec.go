package pubsub

import (
	"context"
	"fmt"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type Codec struct {
	Encoding Encoding

	// DefaultEncodingSelectionFn allows for encoding selection strategies to be injected.
	DefaultEncodingSelectionFn EncodingSelector

	v03 *CodecV03
	v1  *CodecV1
}

const (
	prefix = "ce-"
)

var _ transport.Codec = (*Codec)(nil)

func (c *Codec) Encode(ctx context.Context, e cloudevents.Event) (transport.Message, error) {
	switch c.Encoding {
	case Default, BinaryV03, StructuredV03:
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
		return nil, transport.NewErrMessageEncodingUnknown("wrapper", TransportName)
	}
}

func (c *Codec) Decode(ctx context.Context, msg transport.Message) (*cloudevents.Event, error) {
	switch encoding := c.inspectEncoding(ctx, msg); encoding {
	case BinaryV03, StructuredV03:
		if c.v03 == nil {
			c.v03 = &CodecV03{Encoding: encoding}
		}
		return c.v03.Decode(ctx, msg)
	case BinaryV1, StructuredV1:
		if c.v1 == nil {
			c.v1 = &CodecV1{Encoding: encoding}
		}
		return c.v1.Decode(ctx, msg)
	default:
		return nil, fmt.Errorf("unknown encoding: %s", encoding)
	}
}

func (c *Codec) inspectEncoding(ctx context.Context, msg transport.Message) Encoding {
	if c.v03 == nil {
		c.v03 = &CodecV03{Encoding: c.Encoding}
	}
	// Try v0.3.
	encoding := c.v03.inspectEncoding(ctx, msg)
	if encoding != Unknown {
		return encoding
	}

	// We do not understand the message encoding.
	return Unknown
}
