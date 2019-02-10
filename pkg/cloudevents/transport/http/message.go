package http

import (
	"encoding/json"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"net/http"
)

// type check that this transport message impl matches the contract
var _ transport.Message = (*Message)(nil)

type Message struct {
	Header http.Header
	Body   []byte
}

func (m Message) CloudEventsVersion() string {

	// TODO: the impl of this method needs to move into the codec.

	if m.Header != nil {
		// Try headers first.
		// v0.1
		v := m.Header["CE-CloudEventsVersion"] // TODO: this will fail for real headers.
		if len(v) == 1 {
			return v[0]
		}
		// v0.2
		v = m.Header["ce-SpecVersion"] // TODO: this will fail for real headers.
		if len(v) == 1 {
			return v[0]
		}
	}

	// Then try the data body.
	// TODO: we need to use the correct decoding based on content type.

	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(m.Body, &raw); err != nil {
		return ""
	}

	// v0.1
	if v, ok := raw["cloudEventsVersion"]; ok {
		var version string
		if err := json.Unmarshal(v, &version); err != nil {
			return ""
		}
		return version
	}

	// v0.2
	if v, ok := raw["specVersion"]; ok {
		var version string
		if err := json.Unmarshal(v, &version); err != nil {
			return ""
		}
		return version
	}

	return ""
}
