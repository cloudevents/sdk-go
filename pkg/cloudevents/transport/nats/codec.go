package nats

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
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

// Give the context back as the user expects
func (c *Codec) convertEvent(event *cloudevents.Event) *cloudevents.Event {
	if event == nil {
		return nil
	}
	switch c.Encoding {
	case Default:
		return event
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
