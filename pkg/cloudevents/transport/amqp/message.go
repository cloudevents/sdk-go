package amqp

import (
	"encoding/json"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// type check that this transport message impl matches the contract
var _ transport.Message = (*Message)(nil)

type Message struct {
	ContentType string
	Headers     map[string]interface{}
	Body        []byte
}

// TODO: update this to work with AMQP
func (m Message) CloudEventsVersion() string {
	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(m.Body, &raw); err != nil {
		return ""
	}

	// v0.2
	if v, ok := raw["specversion"]; ok {
		var version string
		if err := json.Unmarshal(v, &version); err != nil {
			return ""
		}
		return version
	}

	return ""
}
