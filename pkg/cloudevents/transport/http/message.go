package http

import (
	"bytes"
	"encoding/json"

	"io"
	"io/ioutil"
	"net/http"

	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// type check that this transport message impl matches the contract
var _ transport.Message = (*Message)(nil)

// Message is an http transport message.
type Message struct {
	Header http.Header
	Body   []byte
}

// Response is an http transport response.
type Response struct {
	StatusCode int
	Message    Message
}

// CloudEventsVersion inspects a message and tries to discover and return the
// CloudEvents spec version.
func (m Message) CloudEventsVersion() string {

	// TODO: the impl of this method needs to move into the codec.

	if m.Header != nil {
		// Try headers first.
		// v0.1, cased from the spec
		if v := m.Header["CE-CloudEventsVersion"]; len(v) == 1 {
			return v[0]
		}
		// v0.2, canonical casing
		if ver := m.Header.Get("CE-CloudEventsVersion"); ver != "" {
			return ver
		}

		// v0.2, cased from the spec
		if v := m.Header["ce-specversion"]; len(v) == 1 {
			return v[0]
		}
		// v0.2, canonical casing
		if ver := m.Header.Get("ce-specversion"); ver != "" {
			return ver
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
	if v, ok := raw["specversion"]; ok {
		var version string
		if err := json.Unmarshal(v, &version); err != nil {
			return ""
		}
		return version
	}

	return ""
}

func readAllClose(r io.ReadCloser) ([]byte, error) {
	if r != nil {
		defer r.Close()
		return ioutil.ReadAll(r)
	}
	return nil, nil
}

// FromRequest initializes a Message from a http.Request
func (m *Message) FromRequest(r *http.Request) (err error) {
	copyHeadersEnsure(r.Header, &m.Header)
	m.Body, err = readAllClose(r.Body)
	return err
}

// ToRequest copies a message to a http.Request.
// Replaces r.Body, r.ContentLength and r.Method.
// Updates r.Headers.
func (m *Message) ToRequest(r *http.Request) {
	copyHeadersEnsure(m.Header, &r.Header)
	if m.Body != nil {
		r.Body = ioutil.NopCloser(bytes.NewBuffer(m.Body))
	} else {
		r.Body = nil
	}
	r.ContentLength = int64(len(m.Body))
	r.Method = http.MethodPost
}
