package http

import (
	"bytes"
	"io"
	nethttp "net/http"
	"strings"

	"github.com/valyala/bytebufferpool"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
)

const prefix = "Ce-"

var specs = spec.WithPrefix(prefix)

const ContentType = "Content-Type"

// Message holds the Header and Body of a HTTP Request.
type Message struct {
	header   nethttp.Header
	body     *bytebufferpool.ByteBuffer
	pool     *bytebufferpool.Pool
	onFinish func(error) error
}

// Check if http.Message implements binding.Message & binding.MessagePayloadReader
var _ binding.Message = (*Message)(nil)
var _ binding.MessagePayloadReader = (*Message)(nil)

// NewMessage returns a Message with header and data from body.
// Reads and closes body.
func NewMessage(pool *bytebufferpool.Pool, header nethttp.Header, body io.ReadCloser) (*Message, error) {
	m := Message{header: header, pool: pool}
	if body != nil {
		m.body = pool.Get()
		_, err := m.body.ReadFrom(body)
		if err != nil {
			return nil, err
		}
		err = body.Close()
		if err != nil {
			return nil, err
		}
	}
	return &m, nil
}

func (m *Message) Structured(encoder binding.StructuredEncoder) error {
	if ft := format.Lookup(m.header.Get(ContentType)); ft == nil {
		return binding.ErrNotStructured
	} else {
		return encoder.SetStructuredEvent(ft, m)
	}
}

func (m *Message) Binary(encoder binding.BinaryEncoder) error {
	version, err := specs.FindVersion(m.header.Get)
	if err != nil {
		return binding.ErrNotBinary
	}

	for k, v := range m.header {
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

	if !m.IsEmpty() {
		return encoder.SetData(m)
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

func (m *Message) IsEmpty() bool {
	return m.body == nil
}

func (m *Message) Bytes() []byte {
	return m.body.Bytes()
}

func (m *Message) Reader() io.Reader {
	return bytes.NewReader(m.body.Bytes())
}

func (m *Message) Finish(err error) error {
	if m.body != nil {
		m.pool.Put(m.body)
	}
	if m.onFinish != nil {
		return m.onFinish(err)
	}
	return nil
}
