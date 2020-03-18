package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudevents/sdk-go/v1/cloudevents"
	"github.com/cloudevents/sdk-go/v1/cloudevents/transport"
)

// CodecStructured represents an structured http transport codec for all versions.
// Intended to be used as a base class.
type CodecStructured struct {
	Encoding Encoding
}

func (v CodecStructured) encodeStructured(ctx context.Context, e cloudevents.Event) (transport.Message, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	msg := &Message{
		Attributes: map[string]string{StructuredContentType: cloudevents.ApplicationCloudEventsJSON},
		Data:       data,
	}

	return msg, nil
}

func (v CodecStructured) decodeStructured(ctx context.Context, version string, msg transport.Message) (*cloudevents.Event, error) {
	m, ok := msg.(*Message)
	if !ok {
		return nil, fmt.Errorf("failed to convert transport.Message to pubsub.Message")
	}
	event := cloudevents.New(version)
	err := json.Unmarshal(m.Data, &event)
	return &event, err
}
