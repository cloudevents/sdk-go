package nats

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type Codec struct {
	Encoding Encoding

	v02 *CodecV02
	v03 *CodecV03
}

var _ transport.Codec = (*Codec)(nil)

func (c *Codec) Encode(e cloudevents.Event) (transport.Message, error) {
	switch c.Encoding {
	case Default:
		fallthrough
	case StructuredV02:
		if c.v02 == nil {
			c.v02 = &CodecV02{Encoding: c.Encoding}
		}
		return c.v02.Encode(e)
	case StructuredV03:
		if c.v03 == nil {
			c.v03 = &CodecV03{Encoding: c.Encoding}
		}
		return c.v03.Encode(e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", c.Encoding)
	}
}

func (c *Codec) Decode(msg transport.Message) (*cloudevents.Event, error) {
	if c.v02 == nil {
		c.v02 = &CodecV02{Encoding: c.Encoding}
	}
	// There is only one encoding as of v0.2
	return c.v02.Decode(msg)
}
