package amqp

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"net/textproto"
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
	case BinaryV02, BinaryV03:
		return c.encodeBinary(e)
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
	switch c.inspectEncoding(msg) {
	case Default:
		fallthrough
	case BinaryV02:
		event := cloudevents.New(cloudevents.CloudEventsVersionV02)
		return c.decodeBinary(msg, &event)
	case BinaryV03:
		event := cloudevents.New(cloudevents.CloudEventsVersionV03)
		return c.decodeBinary(msg, &event)
	case StructuredV02:
		if c.v02 == nil {
			c.v02 = &CodecV02{Encoding: c.Encoding}
		}
		return c.v02.Decode(msg)
	case StructuredV03:
		if c.v03 == nil {
			c.v03 = &CodecV03{Encoding: c.Encoding}
		}
		return c.v03.Decode(msg)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", c.Encoding)
	}
}

func (c Codec) encodeBinary(e cloudevents.Event) (transport.Message, error) {
	headers, err := c.toHeaders(e)
	if err != nil {
		return nil, err
	}
	body, err := e.DataBytes()
	if err != nil {
		return nil, err
	}

	msg := &Message{
		Headers: headers,
		Body:    body,
	}

	if e.DataContentType() != "" {
		msg.ContentType = e.DataContentType()
	} else {
		msg.ContentType = cloudevents.ApplicationJSON
	}

	return msg, nil
}

func (c Codec) toHeaders(e cloudevents.Event) (map[string]interface{}, error) {
	h := make(map[string]interface{})
	h["cloudEvents:specversion"] = e.SpecVersion()
	h["cloudEvents:type"] = e.Type()
	h["cloudEvents:source"] = e.Source()
	h["cloudEvents:id"] = e.ID()
	if !e.Time().IsZero() {
		h["cloudEvents:time"] = e.Time()
	}
	if e.SchemaURL() != "" {
		h["cloudEvents:schemaurl"] = e.SchemaURL()
	}

	// TODO
	//for k, v := range ec.Extensions {
	//	// Per spec, map-valued extensions are converted to a list of headers as:
	//	// CE-attrib-key
	//	if mapVal, ok := v.(map[string]interface{}); ok {
	//		for subkey, subval := range mapVal {
	//			encoded, err := json.Marshal(subval)
	//			if err != nil {
	//				return nil, err
	//			}
	//			h.Set("ce-"+k+"-"+subkey, string(encoded))
	//		}
	//		continue
	//	}
	//	encoded, err := json.Marshal(v)
	//	if err != nil {
	//		return nil, err
	//	}
	//	h.Set("ce-"+k, string(encoded))
	//}

	return h, nil
}

func (c Codec) decodeBinary(msg transport.Message, event *cloudevents.Event) (*cloudevents.Event, error) {
	m, ok := msg.(*Message)
	if !ok {
		return nil, fmt.Errorf("failed to convert transport.Message to amqp.Message")
	}
	err := c.fromHeaders(m.Headers, event)
	if err != nil {
		return nil, err
	}
	var body interface{}
	if len(m.Body) > 0 {
		body = m.Body
	}
	event.Data = body
	event.DataEncoded = true
	return event, nil
}

func (c Codec) fromHeaders(h map[string]interface{}, event *cloudevents.Event) error {
	// Normalize headers.
	for k, v := range h {
		ck := textproto.CanonicalMIMEHeaderKey(k)
		if k != ck {
			delete(h, k)
			h[ck] = v
		}
	}

	ec := event.Context

	if sv, ok := h["cloudEvents:specversion"].(string); ok {
		if err := ec.SetSpecVersion(sv); err != nil {
			return err
		}
	}
	delete(h, "cloudEvents:specversion")

	if id, ok := h["cloudEvents:id"].(string); ok {
		if err := ec.SetID(id); err != nil {
			return err
		}
	}
	delete(h, "cloudEvents:id")

	if t, ok := h["cloudEvents:type"].(string); ok {
		if err := ec.SetType(t); err != nil {
			return err
		}
	}
	delete(h, "cloudEvents:type")

	if s, ok := h["cloudEvents:source"].(string); ok {
		if err := ec.SetSource(s); err != nil {
			return err
		}
	}
	delete(h, "cloudEvents:source")

	if t, ok := h["cloudEvents:time"].(string); ok { // TODO: time can be empty
		timestamp := types.ParseTimestamp(t)
		if err := ec.SetTime(timestamp.Time); err != nil {
			return err
		}
	}
	delete(h, "cloudEvents:time")

	if t, ok := h["cloudEvents:schemaurl"].(string); ok {
		timestamp := types.ParseTimestamp(t)
		if err := ec.SetTime(timestamp.Time); err != nil {
			return err
		}
	}
	delete(h, "cloudEvents:schemaurl")

	// At this point, we have deleted all the known headers.
	// Everything left is assumed to be an extension.

	// TODO
	//extensions := make(map[string]interface{})
	//for k, v := range h {
	//	if len(k) > len("ce-") && strings.EqualFold(k[:len("ce-")], "ce-") {
	//		ak := strings.ToLower(k[len("ce-"):])
	//		if i := strings.Index(ak, "-"); i > 0 {
	//			// attrib-key
	//			attrib := ak[:i]
	//			key := ak[(i + 1):]
	//			if xv, ok := extensions[attrib]; ok {
	//				if m, ok := xv.(map[string]interface{}); ok {
	//					m[key] = v
	//					continue
	//				}
	//				// TODO: revisit how we want to bubble errors up.
	//				return ec, fmt.Errorf("failed to process map type extension")
	//			} else {
	//				m := make(map[string]interface{})
	//				m[key] = v
	//				extensions[attrib] = m
	//			}
	//		} else {
	//			// key
	//			var tmp interface{}
	//			if err := json.Unmarshal([]byte(v[0]), &tmp); err == nil {
	//				extensions[ak] = tmp
	//			} else {
	//				// If we can't unmarshal the data, treat it as a string.
	//				extensions[ak] = v[0]
	//			}
	//		}
	//	}
	//}
	//if len(extensions) > 0 {
	//	ec.Extensions = extensions
	//}
	event.Context = ec
	return nil
}

func (c *Codec) inspectEncoding(msg transport.Message) Encoding {
	if c.v02 == nil {
		c.v02 = &CodecV02{Encoding: c.Encoding}
	}
	// Try v0.2.
	encoding := c.v02.inspectEncoding(msg)
	if encoding != Unknown {
		return encoding
	}

	if c.v03 == nil {
		c.v03 = &CodecV03{Encoding: c.Encoding}
	}
	// Try v0.3.
	encoding = c.v03.inspectEncoding(msg)
	if encoding != Unknown {
		return encoding
	}

	// We do not understand the message encoding.
	return Unknown
}
