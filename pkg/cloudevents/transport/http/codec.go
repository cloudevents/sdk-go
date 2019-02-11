package http

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type Codec struct {
	Encoding Encoding

	v01 *CodecV01
	v02 *CodecV02
}

var _ transport.Codec = (*Codec)(nil)

func (c *Codec) Encode(e cloudevents.Event) (transport.Message, error) {
	switch c.Encoding {
	case Default:
		fallthrough
	case BinaryV01:
		fallthrough
	case StructuredV01:
		if c.v01 == nil {
			c.v01 = &CodecV01{Encoding: c.Encoding}
		}
		return c.v01.Encode(e)
	case BinaryV02:
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
	switch c.inspectEncoding(msg) {
	case BinaryV01:
		fallthrough
	case StructuredV01:
		if c.v01 == nil {
			c.v01 = &CodecV01{Encoding: c.Encoding}
		}
		if event, err := c.v01.Decode(msg); err != nil {
			return nil, err
		} else {
			return c.convertEvent(event), nil
		}
	case BinaryV02:
		fallthrough
	case StructuredV02:
		if c.v02 == nil {
			c.v02 = &CodecV02{Encoding: c.Encoding}
		}
		return c.v02.Decode(msg)
	default:
		return nil, fmt.Errorf("unknown encoding for message %v", msg)
	}
}

// Give the context back as the user expects
func (c *Codec) convertEvent(event *cloudevents.Event) *cloudevents.Event {
	if event == nil {
		return nil
	}
	switch c.Encoding {
	case Default:
		return event
	case BinaryV01:
		fallthrough
	case StructuredV01:
		if c.v01 == nil {
			c.v01 = &CodecV01{Encoding: c.Encoding}
		}
		ctx := event.Context.AsV01()
		event.Context = ctx
		return event
	case BinaryV02:
		fallthrough
	case StructuredV02:
		if c.v02 == nil {
			c.v02 = &CodecV02{Encoding: c.Encoding}
		}
		ctx := event.Context.AsV02()
		event.Context = ctx
		return event
	default:
		return nil
	}
}

func (c *Codec) inspectEncoding(msg transport.Message) Encoding {
	// TODO: there should be a better way to make the version codecs on demand.
	if c.v01 == nil {
		c.v01 = &CodecV01{Encoding: c.Encoding}
	}
	// Try v0.1 first.
	encoding := c.v01.inspectEncoding(msg)
	if encoding != Unknown {
		return encoding
	}

	if c.v02 == nil {
		c.v02 = &CodecV02{Encoding: c.Encoding}
	}
	// Try v0.2.
	encoding = c.v02.inspectEncoding(msg)
	if encoding != Unknown {
		return encoding
	}

	// We do not understand the message encoding.
	return Unknown
}

// ---------
// TODO: Should move these somewhere else. the methods are shared for all versions.

func marshalEvent(event interface{}) ([]byte, error) {
	b, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// ----------
// These go in the data codec:

func isJSONEncoding(encoding string) bool {
	// TODO: this is more tricky, it could be anything +json at the end.
	return encoding == "application/json" || encoding == "text/json"
}

func isXMLEncoding(encoding string) bool {
	return encoding == "application/xml" || encoding == "text/xml"
}

func marshalEventData(encoding string, data interface{}) ([]byte, error) {
	if data == nil {
		return []byte(nil), nil
	}

	var b []byte
	var err error

	if encoding == "" || isJSONEncoding(encoding) {
		b, err = json.Marshal(data)
	} else if isXMLEncoding(encoding) {
		b, err = xml.Marshal(data)
	} else {
		err = fmt.Errorf("cannot encode content type %q", encoding)
	}

	if err != nil {
		return nil, err
	}
	return b, nil
}
