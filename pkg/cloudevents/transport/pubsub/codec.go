package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

type Codec struct {
	Encoding Encoding

	// DefaultEncodingSelectionFn allows for encoding selection strategies to be injected.
	DefaultEncodingSelectionFn EncodingSelector

	v03 *CodecV03
}

const (
	prefix = "ce-"
)

var _ transport.Codec = (*Codec)(nil)

func (c *Codec) Encode(ctx context.Context, e cloudevents.Event) (transport.Message, error) {
	switch c.Encoding {
	case Default:
		fallthrough
	case BinaryV03:
		return c.encodeBinary(ctx, e)
	case StructuredV03:
		if c.v03 == nil {
			c.v03 = &CodecV03{Encoding: c.Encoding}
		}
		return c.v03.Encode(ctx, e)
	default:
		return nil, transport.NewErrMessageEncodingUnknown("wrapper", TransportName)
	}
}

func (c *Codec) Decode(ctx context.Context, msg transport.Message) (*cloudevents.Event, error) {
	switch encoding := c.inspectEncoding(ctx, msg); encoding {
	case BinaryV03:
		event := cloudevents.New(cloudevents.CloudEventsVersionV03)
		return c.decodeBinary(ctx, msg, &event)
	case StructuredV03:
		if c.v03 == nil {
			c.v03 = &CodecV03{Encoding: encoding}
		}
		return c.v03.Decode(ctx, msg)
	default:
		return nil, fmt.Errorf("unknown encoding: %s", encoding)
	}
}

func (c Codec) encodeBinary(ctx context.Context, e cloudevents.Event) (transport.Message, error) {
	attributes, err := c.toAttributes(e)
	if err != nil {
		return nil, err
	}
	data, err := e.DataBytes()
	if err != nil {
		return nil, err
	}

	msg := &Message{
		Attributes: attributes,
		Data:       data,
	}

	return msg, nil
}

func (c Codec) toAttributes(e cloudevents.Event) (map[string]string, error) {
	a := make(map[string]string)
	a[prefix+"specversion"] = e.SpecVersion()
	a[prefix+"type"] = e.Type()
	a[prefix+"source"] = e.Source()
	a[prefix+"id"] = e.ID()
	if !e.Time().IsZero() {
		t := types.Timestamp{Time: e.Time()} // TODO: change e.Time() to return string so I don't have to do this.
		a[prefix+"time"] = t.String()
	}
	if e.DataSchema() != "" {
		a[prefix+"schemaurl"] = e.DataSchema()
	}

	if e.DataContentType() != "" {
		a[prefix+"datacontenttype"] = e.DataContentType()
	} else {
		a[prefix+"datacontenttype"] = cloudevents.ApplicationJSON
	}

	if e.Subject() != "" {
		a[prefix+"subject"] = e.Subject()
	}

	if e.DeprecatedDataContentEncoding() != "" {
		a[prefix+"datacontentencoding"] = e.DeprecatedDataContentEncoding()
	}

	for k, v := range e.Extensions() {
		if mapVal, ok := v.(map[string]interface{}); ok {
			for subkey, subval := range mapVal {
				encoded, err := json.Marshal(subval)
				if err != nil {
					return nil, err
				}
				a[prefix+k+"-"+subkey] = string(encoded)
			}
			continue
		}
		if s, ok := v.(string); ok {
			a[prefix+k] = s
			continue
		}
		encoded, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		a[prefix+k] = string(encoded)
	}

	return a, nil
}

func (c Codec) decodeBinary(ctx context.Context, msg transport.Message, event *cloudevents.Event) (*cloudevents.Event, error) {
	m, ok := msg.(*Message)
	if !ok {
		return nil, fmt.Errorf("failed to convert transport.Message to pubsub.Message")
	}
	err := c.fromAttributes(m.Attributes, event)
	if err != nil {
		return nil, err
	}
	var data interface{}
	if len(m.Data) > 0 {
		data = m.Data
	}
	event.Data = data
	event.DataEncoded = true
	return event, nil
}

func (c Codec) fromAttributes(a map[string]string, event *cloudevents.Event) error {
	// Normalize attributes.
	for k, v := range a {
		ck := strings.ToLower(k)
		if k != ck {
			delete(a, k)
			a[ck] = v
		}
	}

	ec := event.Context

	if sv := a[prefix+"specversion"]; sv != "" {
		if err := ec.SetSpecVersion(sv); err != nil {
			return err
		}
	}
	delete(a, prefix+"specversion")

	if id := a[prefix+"id"]; id != "" {
		if err := ec.SetID(id); err != nil {
			return err
		}
	}
	delete(a, prefix+"id")

	if t := a[prefix+"type"]; t != "" {
		if err := ec.SetType(t); err != nil {
			return err
		}
	}
	delete(a, prefix+"type")

	if s := a[prefix+"source"]; s != "" {
		if err := ec.SetSource(s); err != nil {
			return err
		}
	}
	delete(a, prefix+"source")

	if t := a[prefix+"time"]; t != "" {
		if timestamp, err := types.ParseTimestamp(t); err != nil {
			return err
		} else if err := ec.SetTime(timestamp.Time); err != nil {
			return err
		}
	}
	delete(a, prefix+"time")

	if s := a[prefix+"schemaurl"]; s != "" {
		if err := ec.SetDataSchema(s); err != nil {
			return err
		}
	}
	delete(a, prefix+"schemaurl")

	if s := a[prefix+"subject"]; s != "" {
		if err := ec.SetSubject(s); err != nil {
			return err
		}
	}
	delete(a, prefix+"subject")

	if s := a[prefix+"datacontenttype"]; s != "" {
		if err := ec.SetDataContentType(s); err != nil {
			return err
		}
	}
	delete(a, prefix+"datacontenttype")

	if s := a[prefix+"datacontentencoding"]; s != "" {
		if err := ec.DeprecatedSetDataContentEncoding(s); err != nil {
			return err
		}
	}
	delete(a, prefix+"datacontentencoding")

	// At this point, we have deleted all the known headers.
	// Everything left is assumed to be an extension.

	extensions := make(map[string]interface{})
	for k, v := range a {
		if len(k) > len(prefix) && strings.EqualFold(k[:len(prefix)], prefix) {
			ak := strings.ToLower(k[len(prefix):])
			if i := strings.Index(ak, "-"); i > 0 {
				// attrib-key
				attrib := ak[:i]
				key := ak[(i + 1):]
				if xv, ok := extensions[attrib]; ok {
					if m, ok := xv.(map[string]interface{}); ok {
						m[key] = v
						continue
					}
					// TODO: revisit how we want to bubble errors up.
					return fmt.Errorf("failed to process map type extension")
				} else {
					m := make(map[string]interface{})
					m[key] = v
					extensions[attrib] = m
				}
			} else {
				// key
				var tmp interface{}
				if err := json.Unmarshal([]byte(v), &tmp); err == nil {
					extensions[ak] = tmp
				} else {
					// If we can't unmarshal the data, treat it as a string.
					extensions[ak] = v
				}
			}
		}
	}
	event.Context = ec
	if len(extensions) > 0 {
		for k, v := range extensions {
			event.SetExtension(k, v)
		}
	}
	return nil
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
