package binding

import (
	"bytes"
	"io"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
)

// MockBinaryMessage implements a binary-mode message as a simple struct.
type MockBinaryMessage struct {
	Metadata   map[spec.Attribute]interface{}
	Extensions map[string]interface{}
	Body       []byte
}

var versions = spec.New()

func NewMockBinaryMessage(e cloudevents.Event) *MockBinaryMessage {
	version, err := versions.Version(e.SpecVersion())
	if err != nil {
		panic(err)
	}

	m := MockBinaryMessage{
		Metadata:   make(map[spec.Attribute]interface{}),
		Extensions: make(map[string]interface{}),
	}

	for _, attribute := range version.Attributes() {
		val := attribute.Get(e.Context)
		if val != nil {
			m.Metadata[attribute] = val
		}
	}

	for k, v := range e.Extensions() {
		m.Extensions[k] = v
	}

	m.Body, err = e.DataBytes()
	if err != nil {
		panic(err)
	}

	return &m
}

func (bm *MockBinaryMessage) Event(b EventEncoder) error {
	e, _, _, err := ToEvent(bm)
	if err != nil {
		return err
	}
	return b.SetEvent(e)
}

func (bm *MockBinaryMessage) Structured(b StructuredEncoder) error {
	return ErrNotStructured
}

func (bm *MockBinaryMessage) Binary(b BinaryEncoder) error {
	for k, v := range bm.Metadata {
		err := b.SetAttribute(k, v)
		if err != nil {
			return err
		}
	}
	for k, v := range bm.Extensions {
		err := b.SetExtension(k, v)
		if err != nil {
			return err
		}
	}
	if len(bm.Body) == 0 {
		return nil
	}
	return b.SetData(bm)
}

func (bm *MockBinaryMessage) IsEmpty() bool {
	return bm.Body == nil
}

func (bm *MockBinaryMessage) Bytes() []byte {
	return bm.Body
}

func (bm *MockBinaryMessage) Reader() io.Reader {
	return bytes.NewReader(bm.Body)
}

func (bm *MockBinaryMessage) Finish(error) error { return nil }

var _ Message = (*MockBinaryMessage)(nil) // Test it conforms to the interface
