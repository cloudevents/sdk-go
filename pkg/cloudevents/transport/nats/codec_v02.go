package nats

import (
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type CodecV02 struct {
	Encoding Encoding
}

var _ transport.Codec = (*CodecV02)(nil)

func (v CodecV02) Encode(e cloudevents.Event) (transport.Message, error) {
	switch v.Encoding {
	case Default:
		fallthrough
	case StructuredV02:
		return v.encodeStructured(e)
	default:
		return nil, fmt.Errorf("unknown encoding: %d", v.Encoding)
	}
}

func (v CodecV02) Decode(msg transport.Message) (*cloudevents.Event, error) {
	// only structured is supported as of v0.2
	return v.decodeStructured(msg)
}

func (v CodecV02) encodeStructured(e cloudevents.Event) (transport.Message, error) {
	ctxv2 := e.Context.AsV02()
	if ctxv2.ContentType == "" {
		ctxv2.ContentType = "application/json"
	}

	ctx, err := marshalEvent(ctxv2)
	if err != nil {
		return nil, err
	}

	var body []byte

	b := map[string]json.RawMessage{}
	if err := json.Unmarshal([]byte(ctx), &b); err != nil {
		return nil, err
	}

	data, err := marshalEventData(e.Context.DataContentType(), e.Data)
	if err != nil {
		return nil, err
	}
	if data != nil {
		b["data"] = data
	}

	body, err = json.Marshal(b)
	if err != nil {
		return nil, err
	}

	msg := &Message{
		Body: body,
	}

	return msg, nil
}

func (v CodecV02) decodeStructured(msg transport.Message) (*cloudevents.Event, error) {
	m, ok := msg.(*Message)
	if !ok {
		return nil, fmt.Errorf("failed to convert transport.Message to http.Message")
	}

	ec := cloudevents.EventContextV02{}
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

	return &cloudevents.Event{
		Context: ec,
		Data:    data,
	}, nil
}
