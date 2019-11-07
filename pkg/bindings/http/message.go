package http

import (
	"io"
	"io/ioutil"
	nethttp "net/http"

	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

const prefix = "ce-"

var specs = spec.WithPrefix(prefix)

const ContentType = "Content-Type"

// Message holds the Header and Body of a HTTP Request or Response.
type Message struct {
	Header   nethttp.Header
	Body     []byte
	OnFinish func(error) error
}

// NewMessage returns a Message with header and data from body.
// Reads and closes body.
func NewMessage(header nethttp.Header, body io.ReadCloser) (*Message, error) {
	m := Message{Header: header}
	if body != nil {
		defer func() { _ = body.Close() }()
		var err error
		if m.Body, err = ioutil.ReadAll(body); err != nil && err != io.EOF {
			return nil, err
		}
		if len(m.Body) == 0 {
			m.Body = nil
		}
	}
	return &m, nil
}

func (m *Message) Structured() (string, []byte) {
	if ct := m.Header.Get(ContentType); format.IsFormat(ct) {
		return ct, m.Body
	}
	return "", nil
}

func (m *Message) Event() (e ce.Event, err error) {
	if f, b := m.Structured(); f != "" {
		err := format.Unmarshal(f, b, &e)
		return e, err
	}
	version, err := specs.FindVersion(m.Header.Get)
	if err != nil {
		return e, err
	}
	c := version.NewContext()
	if err := c.SetDataContentType(m.Header.Get(ContentType)); err != nil {
		return e, err
	}
	for k, v := range m.Header {
		if err := version.SetAttribute(c, k, v[0]); err != nil {
			return e, err
		}
	}
	if len(m.Body) == 0 {
		return ce.Event{Data: nil, Context: c}, nil
	}
	return ce.Event{Data: m.Body, DataEncoded: true, Context: c}, nil
}

func (m *Message) Finish(err error) error {
	if m.OnFinish != nil {
		return m.OnFinish(err)
	}
	return nil
}
