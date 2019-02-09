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

func (m Message) CloudEventVersion() string {

	// TODO: the impl of this method needs to move into the codec.

	if m.Header != nil {
		// Try headers first.
		// v0.1
		v := m.Header["CE-CloudEventsVersion"]
		if len(v) == 1 {
			return v[1]
		}
		// v0.2
		v = m.Header["ce-SpecVersion"]
		if len(v) == 1 {
			return v[1]
		}
	}

	// Then try the data body.
	b := map[string]string{}
	if err := json.Unmarshal(m.Body, &b); err != nil {
		return ""
	}

	// v0.1
	if b["cloudEventsVersion"] != "" {
		return b["cloudEventsVersion"]
	}

	// v0.2
	if b["specVersion"] != "" {
		return b["specVersion"]
	}

	return ""
}
