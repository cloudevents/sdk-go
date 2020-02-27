package nats

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/nats-io/nats.go"
)

type Message struct {
	*nats.Msg
}

var _ binding.Message = (*Message)(nil)

// By the specification, all NATS message payloads must be in the JSON event format.
// If this isn't the case then this function will return an error
func NewMessage(msg *nats.Msg) (*Message, error) {
	raw := make(map[string]json.RawMessage)

	if err := json.Unmarshal(msg.Data, &raw); err != nil {
		return nil, err
	}

	if _, err := spec.VS.FindVersion(versionExtractor(raw)); err != nil {
		return nil, err
	}

	return &Message{Msg: msg}, nil
}

// All NATS messages are structured
func (m *Message) Encoding() binding.Encoding {
	return binding.EncodingStructured
}

func (m *Message) Structured(ctx context.Context, encoder binding.StructuredEncoder) error {
	return encoder.SetStructuredEvent(ctx, format.JSON, bytes.NewReader(m.Data))
}

func (m *Message) Binary(context.Context, binding.BinaryEncoder) error {
	return BinaryEncodingNotSupported
}

func (m *Message) Finish(error) error {
	return nil
}

// Reads versionKey from the raw map, and if the value exists and is a string returns it, otherwise returns ""
func versionExtractor(raw map[string]json.RawMessage) func(string) string {
	return func(versionKey string) string {
		if v := raw[versionKey]; v != nil {
			var version string
			if err := json.Unmarshal(v, &version); err == nil {
				return version
			}
		}
		return ""
	}
}
