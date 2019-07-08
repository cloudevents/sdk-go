package amqp

import (
	"encoding/json"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/transport"
)

// CodecStructured represents an structured http transport codec for all versions.
// Intended to be used as a base class.
type CodecStructured struct {
	Encoding Encoding
}

func (v CodecStructured) encodeStructured(e cloudevents.Event) (transport.Message, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	msg := &Message{
		ContentType: cloudevents.ApplicationCloudEventsJSON,
		Body:        body,
	}

	return msg, nil
}

func (v CodecStructured) decodeStructured(version string, msg transport.Message) (*cloudevents.Event, error) {
	m, ok := msg.(*Message)
	if !ok {
		return nil, fmt.Errorf("failed to convert transport.Message to amqp.Message")
	}
	event := cloudevents.NewEvent(version)
	err := json.Unmarshal(m.Body, &event)
	return &event, err
}
