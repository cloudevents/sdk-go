package http

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"net/http"
	"strings"
)

type CodecV01 struct {
	Encoding Encoding
}

var _ transport.Codec = (*CodecV01)(nil)

func (v CodecV01) Encode(e canonical.Event) (transport.Message, error) {
	switch v.Encoding {
	case Default:
		fallthrough
	case BinaryV01:
		return v.encodeBinary(e)
	case StructuredV01:
		return v.encodeStructured(e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", v.Encoding)
	}
}

func (v CodecV01) Decode(msg transport.Message) (*canonical.Event, error) {
	switch v.inspectEncoding(msg) {
	case BinaryV01:
		return v.decodeBinary(msg)
	case StructuredV01:
		return v.decodeStructured(msg)
	default:
		return nil, fmt.Errorf("unknown encoding for message %v", msg)
	}
}

func (v CodecV01) encodeBinary(e canonical.Event) (transport.Message, error) {
	header, err := v.asHeaders(e.Context.AsV01())
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

func (v CodecV01) asHeaders(ec canonical.EventContextV01) (http.Header, error) {
	// Preserve case in v0.1, even though HTTP headers are case-insensitive.
	h := http.Header{}
	h["CE-CloudEventsVersion"] = []string{ec.CloudEventsVersion}
	h["CE-EventID"] = []string{ec.EventID}
	h["CE-EventType"] = []string{ec.EventType}
	h["CE-Source"] = []string{ec.Source.String()}
	if ec.EventTime != nil && !ec.EventTime.IsZero() {
		h["CE-EventTime"] = []string{ec.EventTime.String()}
	}
	if ec.EventTypeVersion != "" {
		h["CE-EventTypeVersion"] = []string{ec.EventTypeVersion}
	}
	if ec.SchemaURL != nil {
		h["CE-SchemaURL"] = []string{ec.SchemaURL.String()}
	}
	if ec.ContentType != "" {
		h.Set("Content-Type", ec.ContentType)
	} else if v.Encoding == Default || v.Encoding == BinaryV01 {
		// in binary v0.1, the Content-Type header is tied to ec.ContentType
		// This was later found to be an issue with the spec, but yolo.
		// TODO: not sure what the default should be?
		h.Set("Content-Type", "application/json")
	}

	// Regarding Extensions, v0.1 Spec says the following:
	// * Each map entry name MUST be prefixed with "CE-X-"
	// * Each map entry name's first character MUST be capitalized
	for k, v := range ec.Extensions {
		encoded, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		h["CE-X-"+strings.Title(k)] = []string{string(encoded)}
	}
	return h, nil
}

func (v CodecV01) encodeStructured(e canonical.Event) (transport.Message, error) {
	header := http.Header{}
	header.Set("Content-Type", "application/cloudevents+json")

	ctx, err := marshalEvent(e.Context.AsV01())
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

func (v CodecV01) decodeBinary(msg transport.Message) (*canonical.Event, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v CodecV01) decodeStructured(msg transport.Message) (*canonical.Event, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v CodecV01) inspectEncoding(msg transport.Message) Encoding {
	return Unknown
}

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
