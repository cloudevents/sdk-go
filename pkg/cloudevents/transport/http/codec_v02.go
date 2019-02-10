package http

import (
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"log"
	"net/http"
	"net/textproto"
)

type CodecV02 struct {
	Encoding Encoding
}

var _ transport.Codec = (*CodecV02)(nil)

func (v CodecV02) Encode(e canonical.Event) (transport.Message, error) {
	switch v.Encoding {
	case Default:
		fallthrough
	case BinaryV02:
		return v.encodeBinary(e)
	case StructuredV02:
		return v.encodeStructured(e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", v.Encoding)
	}
}

func (v CodecV02) Decode(msg transport.Message) (*canonical.Event, error) {
	switch v.inspectEncoding(msg) {
	case BinaryV02:
		return v.decodeBinary(msg)
	case StructuredV02:
		return v.decodeStructured(msg)
	default:
		return nil, fmt.Errorf("unknown encoding for message %v", msg)
	}
}

func (v CodecV02) encodeBinary(e canonical.Event) (transport.Message, error) {
	header, err := v.toHeaders(e.Context.AsV02())
	if err != nil {
		return nil, err
	}

	body, err := marshalEventData(e.Context.DataContentType(), e.Data)
	if err != nil {
		return nil, err
	}

	msg := &Message{
		Header: header,
		Body:   body,
	}

	return msg, nil
}

func (v CodecV02) toHeaders(ec canonical.EventContextV02) (http.Header, error) {
	h := http.Header{}
	h.Set("ce-specversion", ec.SpecVersion)
	h.Set("ce-type", ec.Type)
	h.Set("ce-source", ec.Source.String())
	h.Set("ce-id", ec.ID)
	if ec.Time != nil && !ec.Time.IsZero() {
		h.Set("ce-time", ec.Time.String())
	}
	if ec.SchemaURL != nil {
		h.Set("ce-schemaurl", ec.SchemaURL.String())
	}
	if ec.ContentType != "" {
		h.Set("Content-Type", ec.ContentType)
	} else if v.Encoding == Default || v.Encoding == BinaryV02 {
		// in binary v0.2, the Content-Type header is tied to ec.ContentType
		// This was later found to be an issue with the spec, but yolo.
		// TODO: not sure what the default should be?
		h.Set("Content-Type", "application/json")
	}
	for k, v := range ec.Extensions {
		// Per spec, map-valued extensions are converted to a list of headers as:
		// CE-attrib-key
		if mapVal, ok := v.(map[string]interface{}); ok {
			for subkey, subval := range mapVal {
				encoded, err := json.Marshal(subval)
				if err != nil {
					return nil, err
				}
				h.Set("ce-"+k+"-"+subkey, string(encoded))
			}
			continue
		}
		encoded, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		h.Set("ce-"+k, string(encoded))
	}

	return h, nil
}

func (v CodecV02) encodeStructured(e canonical.Event) (transport.Message, error) {
	header := http.Header{}
	header.Set("Content-Type", "application/cloudevents+json")

	ctx, err := marshalEvent(e.Context.AsV02())
	if err != nil {
		return nil, err
	}

	var body []byte

	b := map[string]interface{}{}
	if err := json.Unmarshal([]byte(ctx), &b); err != nil {
		return nil, err
	}

	dataContentType := e.Context.DataContentType()
	if dataContentType == "application/json" {
		if e.Data != nil {
			b["data"] = e.Data
		}
	} else {
		data, err := marshalEventData(e.Context.DataContentType(), e.Data)
		if err != nil {
			return nil, err
		}

		if data != nil {
			b["data"] = data
		}
	}
	body, err = json.Marshal(b)
	if err != nil {
		return nil, err
	}

	msg := &Message{
		Header: header,
		Body:   body,
	}

	return msg, nil
}

func (v CodecV02) decodeBinary(msg transport.Message) (*canonical.Event, error) {
	m, ok := msg.(*Message)
	if !ok {
		return nil, fmt.Errorf("failed to convert transport.Message to http.Message")
	}
	ctx, err := v.fromHeaders(m.Header)
	if err != nil {
		return nil, err
	}
	var body interface{}
	if len(m.Body) > 0 {
		body = m.Body
	}
	return &canonical.Event{
		Context: ctx,
		Data:    body,
	}, nil
}

func (v CodecV02) fromHeaders(h http.Header) (canonical.EventContextV02, error) {
	// Normalize headers.
	for k, v := range h {
		ck := textproto.CanonicalMIMEHeaderKey(k)
		if k != ck {
			log.Printf("[warn] received header with non-canonical form; canonical: %q, got %q", ck, k)
			h[ck] = v
		}
	}

	ec := canonical.EventContextV02{}
	ec.SpecVersion = h.Get("ce-specversion")
	ec.ID = h.Get("ce-id")
	ec.Type = h.Get("ce-type")
	source := canonical.ParseURLRef(h.Get("ce-source"))
	if source != nil {
		ec.Source = *source
	}
	ec.Time = canonical.ParseTimestamp(h.Get("ce-time"))
	ec.SchemaURL = canonical.ParseURLRef(h.Get("ce-schemaurl"))
	ec.ContentType = h.Get("Content-Type")

	// TODO: fix extensions
	//extensions := make(map[string]interface{})
	//for k, v := range h {
	//	if strings.EqualFold(k[:len("CE-X-")], "CE-X-") {
	//		key := k[len("CE-X-"):]
	//		var tmp interface{}
	//		if err := json.Unmarshal([]byte(v[0]), &tmp); err == nil {
	//			extensions[key] = tmp
	//		} else {
	//			// If we can't unmarshal the data, treat it as a string.
	//			extensions[key] = v[0]
	//		}
	//	}
	//}
	//if len(extensions) > 0 {
	//	ec.Extensions = extensions
	//}
	return ec, nil
}

func (v CodecV02) decodeStructured(msg transport.Message) (*canonical.Event, error) {
	m, ok := msg.(*Message)
	if !ok {
		return nil, fmt.Errorf("failed to convert transport.Message to http.Message")
	}

	ec := canonical.EventContextV02{}
	if err := json.Unmarshal(m.Body, &ec); err != nil {
		return nil, err
	}

	raw := make(map[string]json.RawMessage)

	if err := json.Unmarshal(m.Body, &raw); err != nil {
		return nil, err
	}
	var data interface{}
	if d, ok := raw["data"]; ok {
		data = []byte(d)
	}

	return &canonical.Event{
		Context: ec,
		Data:    data,
	}, nil
}

func (v CodecV02) inspectEncoding(msg transport.Message) Encoding {
	version := msg.CloudEventsVersion()
	if version != canonical.CloudEventsVersionV02 {
		return Unknown
	}
	m, ok := msg.(*Message)
	if !ok {
		return Unknown
	}
	contentType := m.Header.Get("Content-Type")
	if contentType == "application/json" {
		return BinaryV02
	}
	if contentType == "application/cloudevents+json" {
		return StructuredV02
	}
	return Unknown
}
