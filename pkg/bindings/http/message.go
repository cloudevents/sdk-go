package http

import (
	"context"
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
	format     format.Format
	version    spec.Version
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
	if m.format = format.Lookup(header.Get(ContentType)); m.format == nil {
		m.version, _ = specs.FindVersion(m.Header.Get)
	}
	return &m, nil
}

func (m *Message) Encoding() binding.Encoding {
	if m.version != nil {
		return binding.EncodingBinary
	}
	if m.format != nil {
		return binding.EncodingStructured
	}
	return binding.EncodingUnknown
}

func (m *Message) Structured(ctx context.Context, encoder binding.StructuredEncoder) error {
	if m.format == nil {
		return binding.ErrNotStructured
	} else {
		return encoder.SetStructuredEvent(ctx, m.format, m.BodyReader)
	}
}

func (m *Message) Binary(ctx context.Context, encoder binding.BinaryEncoder) error {
	if m.version == nil {
		return binding.ErrNotBinary
	}

	err := encoder.Start(ctx)
	if err != nil {
		return err
	}

	for k, v := range m.Header {
		if strings.HasPrefix(k, prefix) {
			attr := m.version.Attribute(k)
			if attr != nil {
				err = encoder.SetAttribute(attr, v[0])
			} else {
				err = encoder.SetExtension(strings.ToLower(strings.TrimPrefix(k, prefix)), v[0])
			}
		} else if k == ContentType {
			err = encoder.SetAttribute(m.version.AttributeFromKind(spec.DataContentType), v[0])
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

	return encoder.End()
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
