package pubsub

import (
	"encoding/json"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// type check that this transport message impl matches the contract
var _ transport.Message = (*Message)(nil)

// Message represents a Pub/Sub message.
type Message struct {
	// Data is the actual data in the message.
	Data []byte

	// Attributes represents the key-value pairs the current message
	// is labelled with.
	Attributes map[string]string
}

func (m Message) CloudEventsVersion() string {
	// Check as Binary encoding first.
	if m.Attributes != nil {
		// Binary v0.3:
		if s := m.Attributes[prefix+"specversion"]; s != "" {
			return s
		}
	}

	// Now check as Structured encoding.
	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(m.Data, &raw); err != nil {
		return ""
	}

	// structured v0.3
	if v, ok := raw["specversion"]; ok {
		var version string
		if err := json.Unmarshal(v, &version); err != nil {
			return ""
		}
		return version
	}

	return ""
}
