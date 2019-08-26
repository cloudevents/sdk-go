package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// CodecStructured represents an structured http transport codec for all versions.
// Intended to be used as a base class.
type CodecStructured struct {
	Encoding Encoding
}

func (v CodecStructured) encodeStructured(ctx context.Context, e cloudevents.Event) (transport.Message, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	msg := &Message{
		Body: body,
	}

	return msg, nil
}

func (v CodecStructured) decodeStructured(ctx context.Context, version string, msg transport.Message) (*cloudevents.Event, error) {
	m, ok := msg.(*Message)
	if !ok {
		return nil, fmt.Errorf("failed to convert transport.Message to nats.Message")
	}
	event := cloudevents.New(version)
	err := json.Unmarshal(m.Body, &event)
	return &event, err
}
