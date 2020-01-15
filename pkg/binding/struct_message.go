package binding

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/cloudevents/sdk-go/pkg/binding/format"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

// StructMessage implements a structured-mode message as a simple struct.
type StructMessage struct {
	Format format.Format
	Bytes  []byte
}

// Event unmarshals the formatted event.
func (m *StructMessage) Event(enc EventEncoder) error {
	var e ce.Event
	err := m.Format.Unmarshal(m.Bytes, &e)
	if err != nil {
		return err
	}
	return enc.SetEvent(e)
}

// Structured copies structured data to a StructuredEncoder
func (m *StructMessage) Structured(enc StructuredEncoder) error {
	return enc.SetStructuredEvent(m.Format, bytes.NewReader(m.Bytes))
}

// Binary returns ErrNotBinary
func (m StructMessage) Binary(enc BinaryEncoder) error { return ErrNotBinary }

func (*StructMessage) Finish(error) error { return nil }

func (m *StructMessage) SetStructuredEvent(format format.Format, event io.Reader) error {
	b, err := ioutil.ReadAll(event)
	if err != nil {
		return err
	}
	m.Format = format
	m.Bytes = b
	return nil
}

var _ Message = (*StructMessage)(nil)           // Test it conforms to the interface
var _ StructuredEncoder = (*StructMessage)(nil) // Test it conforms to the interface
