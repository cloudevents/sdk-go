package http

import (
	"io"
	nethttp "net/http"
	"strings"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
)

const prefix = "Ce-"

var specs = spec.WithPrefix(prefix)

const ContentType = "Content-Type"

// Message holds the Header and Body of a HTTP Request or Response.
type Message struct {
	Header     nethttp.Header
	BodyReader io.ReadCloser
	OnFinish   func(error) error
}

// Check if http.Message implements binding.Message
var _ binding.Message = (*Message)(nil)

// NewMessage returns a Message with header and data from body.
// Reads and closes body.
func NewMessage(header nethttp.Header, body io.ReadCloser) (*Message, error) {
	m := Message{Header: header}
	if body != nil {
		m.BodyReader = body
	}
	return &m, nil
}

func (m *Message) Structured(encoder binding.StructuredEncoder) error {
	if ft := format.Lookup(m.Header.Get(ContentType)); ft == nil {
		return binding.ErrNotStructured
	} else {
		return encoder.SetStructuredEvent(ft, m.BodyReader)
	}
}

func (m *Message) Binary(encoder binding.BinaryEncoder) error {
	version, err := specs.FindVersion(m.Header.Get)
	if err != nil {
		return binding.ErrNotBinary
	}

	for k, v := range m.Header {
		if strings.HasPrefix(k, prefix) {
			attr := version.Attribute(k)
			if attr != nil {
				err = encoder.SetAttribute(attr, v[0])
			} else {
				err = encoder.SetExtension(strings.ToLower(strings.TrimPrefix(k, prefix)), v[0])
			}
		} else if k == ContentType {
			err = encoder.SetAttribute(version.AttributeFromKind(spec.DataContentType), v[0])
		}
		if err != nil {
			return err
		}
	}

	if m.BodyReader != nil {
		err = encoder.SetData(m.BodyReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Message) Event(encoder binding.EventEncoder) error {
	e, _, _, err := binding.ToEvent(m)
	if err != nil {
		return err
	}
	return encoder.SetEvent(e)
}

func (m *Message) Finish(err error) error {
	if m.BodyReader != nil {
		_ = m.BodyReader.Close()
	}
	if m.OnFinish != nil {
		return m.OnFinish(err)
	}
	return nil
}
