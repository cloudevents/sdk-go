package http

import (
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"net/http"
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
	return nil, fmt.Errorf("not implemented")
}

func (v CodecV02) decodeStructured(msg transport.Message) (*canonical.Event, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v CodecV02) inspectEncoding(msg transport.Message) Encoding {
	return Unknown
}
