package http

import (
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"log"
	"net/http"
	"net/textproto"
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
	header, err := v.toHeaders(e.Context.AsV01())
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

func (v CodecV01) toHeaders(ec canonical.EventContextV01) (http.Header, error) {
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

func (v CodecV01) fromHeaders(h http.Header) (canonical.EventContextV01, error) {

	// Normalize headers.
	for k, v := range h {
		ck := textproto.CanonicalMIMEHeaderKey(k)
		if k != ck {
			log.Printf("[warn] received header with non-canonical form; canonical: %q, got %q", ck, k)
			h[ck] = v
		}
	}

	ec := canonical.EventContextV01{}
	ec.CloudEventsVersion = h.Get("CE-CloudEventsVersion")
	ec.EventID = h.Get("CE-EventID")
	ec.EventType = h.Get("CE-EventType")
	source := canonical.ParseURLRef(h.Get("CE-Source"))
	if source != nil {
		ec.Source = *source
	}
	ec.EventTime = canonical.ParseTimestamp(h.Get("CE-EventTime"))
	ec.EventTypeVersion = h.Get("CE-EventTypeVersion")
	ec.SchemaURL = canonical.ParseURLRef(h.Get("CE-SchemaURL"))
	ec.ContentType = h.Get("Content-Type")

	//for k, v := range ec.Extensions {  <-- this is a challenge
	//	encoded, err := json.Marshal(v)
	//	if err != nil {
	//		return nil, err
	//	}
	//	h["CE-X-"+strings.Title(k)] = []string{string(encoded)}
	//}
	return ec, nil
}

func (v CodecV01) decodeStructured(msg transport.Message) (*canonical.Event, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v CodecV01) inspectEncoding(msg transport.Message) Encoding {
	return BinaryV01 // Lets try to get the decoder working before inspection code.
	//return Unknown
}
