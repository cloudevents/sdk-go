package pubsub

import (
	"encoding/json"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// type check that this transport message impl matches the contract
var _ transport.Message = (*Message)(nil)

type Message struct {
	Attributes map[string]string
	Body       []byte
}

func (m Message) CloudEventsVersion() string {
	// Check as Binary encoding first.
	if m.Attributes != nil {
		// Binary v0.2, v0.3:
		if s := m.Attributes[prefix+"specversion"]; s != "" {
			return s
		}
	}

	// Now check as Structured encoding.
	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(m.Body, &raw); err != nil {
		return ""
	}

	// structured v0.2, v0.3
	if v, ok := raw["specversion"]; ok {
		var version string
		if err := json.Unmarshal(v, &version); err != nil {
			return ""
		}
		return version
	}

	return ""
}
