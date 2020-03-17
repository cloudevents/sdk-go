package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudevents/sdk-go/legacy/pkg/cloudevents"
	cecontext "github.com/cloudevents/sdk-go/legacy/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/legacy/pkg/cloudevents/transport"
	"github.com/cloudevents/sdk-go/legacy/pkg/cloudevents/types"
)

const (
	StructuredContentType = "Content-Type"
)

type CodecV03 struct {
	CodecStructured

	DefaultEncoding Encoding
}

var _ transport.Codec = (*CodecV03)(nil)

func (v CodecV03) Encode(ctx context.Context, e cloudevents.Event) (transport.Message, error) {
	encoding := v.DefaultEncoding
	strEnc := cecontext.EncodingFrom(ctx)
	if strEnc != "" {
		switch strEnc {
		case Binary:
			encoding = BinaryV03
		case Structured:
			encoding = StructuredV03
		}
	}
	switch encoding {
	case Default:
		fallthrough
	case StructuredV03:
		return v.encodeStructured(ctx, e)
	case BinaryV03:
		return v.encodeBinary(ctx, e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", encoding)
	}
}

func (v CodecV03) Decode(ctx context.Context, msg transport.Message) (*cloudevents.Event, error) {
	// only structured is supported as of v0.3
	switch v.inspectEncoding(ctx, msg) {
	case StructuredV03:
		return v.decodeStructured(ctx, cloudevents.CloudEventsVersionV03, msg)
	case BinaryV03:
		event := cloudevents.New(cloudevents.CloudEventsVersionV03)
		return v.decodeBinary(ctx, msg, &event)
	default:
		return nil, transport.NewErrMessageEncodingUnknown("v03", TransportName)
	}
}

func (v CodecV03) inspectEncoding(ctx context.Context, msg transport.Message) Encoding {
	version := msg.CloudEventsVersion()
	if version != cloudevents.CloudEventsVersionV03 {
		return Unknown
	}
	m, ok := msg.(*Message)
	if !ok {
		return Unknown
	}
	if m.Attributes[StructuredContentType] == cloudevents.ApplicationCloudEventsJSON {
		return StructuredV03
	}
	return BinaryV03
}

func (v CodecV03) encodeBinary(ctx context.Context, e cloudevents.Event) (transport.Message, error) {
	attributes, err := v.toAttributes(e)
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

func (v CodecV03) toAttributes(e cloudevents.Event) (map[string]string, error) {
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

func (v CodecV03) decodeBinary(ctx context.Context, msg transport.Message, event *cloudevents.Event) (*cloudevents.Event, error) {
	m, ok := msg.(*Message)
	if !ok {
		return nil, fmt.Errorf("failed to convert transport.Message to pubsub.Message")
	}
	err := v.fromAttributes(m.Attributes, event)
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

func (v CodecV03) fromAttributes(a map[string]string, event *cloudevents.Event) error {
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
