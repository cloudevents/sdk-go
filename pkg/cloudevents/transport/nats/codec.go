package nats

import (
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/datacodec"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"log"
)

type Codec struct {
	Encoding Encoding

	v02 *CodecV02
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

// ---------
// TODO: Should move these somewhere else. the methods are shared for all versions.

func marshalEvent(event interface{}) ([]byte, error) {

	if b, ok := event.([]byte); ok {
		log.Printf("json.marshalEvent asked to encode bytes... wrong? %s", string(b))
	}

	b, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func marshalEventData(encoding string, data interface{}) ([]byte, error) {
	if data == nil {
		return []byte(nil), nil
	}
	// already encoded?
	if b, ok := data.([]byte); ok {
		return b, nil
	}
	return datacodec.Encode(encoding, data)
}
