package buffering

import (
	"bytes"
	"io"

	"github.com/valyala/bytebufferpool"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

var structMessagePool bytebufferpool.Pool

// structBufferedMessage implements a structured-mode message as a simple struct.
// This message implementation is used by CopyMessage and BufferMessage
type structBufferedMessage struct {
	Format format.Format
	Bytes  *bytebufferpool.ByteBuffer
}

// Event unmarshals the formatted event.
func (m *structBufferedMessage) Event(enc binding.EventEncoder) error {
	var e ce.Event
	err := m.Format.Unmarshal(m.Bytes.B, &e)
	if err != nil {
		return err
	}
	return enc.SetEvent(e)
}

// Structured copies structured data to a StructuredEncoder
func (m *structBufferedMessage) Structured(enc binding.StructuredEncoder) error {
	return enc.SetStructuredEvent(m.Format, bytes.NewReader(m.Bytes.B))
}

// Binary returns ErrNotBinary
func (m structBufferedMessage) Binary(enc binding.BinaryEncoder) error { return binding.ErrNotBinary }

func (m *structBufferedMessage) Finish(error) error {
	structMessagePool.Put(m.Bytes)
	return nil
}

func (m *structBufferedMessage) SetStructuredEvent(format format.Format, event io.Reader) error {
	m.Bytes = structMessagePool.Get()
	_, err := io.Copy(m.Bytes, event)
	if err != nil {
		return err
	}
	m.Format = format
	return nil
}

var _ binding.Message = (*structBufferedMessage)(nil)           // Test it conforms to the interface
var _ binding.StructuredEncoder = (*structBufferedMessage)(nil) // Test it conforms to the interface
