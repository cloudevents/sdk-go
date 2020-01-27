package buffering

import (
	"bytes"
	"io"

	"github.com/valyala/bytebufferpool"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/event"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
)

var binaryMessagePool bytebufferpool.Pool

// binaryBufferedMessage implements a binary-mode message as a simple struct.
// This message implementation is used by CopyMessage and BufferMessage
type binaryBufferedMessage struct {
	metadata   map[spec.Attribute]interface{}
	extensions map[string]interface{}
	body       *bytebufferpool.ByteBuffer
}

func (m *binaryBufferedMessage) Structured(binding.StructuredEncoder) error {
	return binding.ErrNotStructured
}

func (m *binaryBufferedMessage) Binary(b binding.BinaryEncoder) (err error) {
	for k, v := range m.metadata {
		err := b.SetAttribute(k, v)
		if err != nil {
			return err
		}
	}
	for k, v := range m.extensions {
		err := b.SetExtension(k, v)
		if err != nil {
			return err
		}
	}
	if m.body != nil {
		return b.SetData(bytes.NewReader(m.body.Bytes()))
	}
	return err
}

func (m *binaryBufferedMessage) Event(builder binding.EventEncoder) error {
	e, _, _, err := event.ToEvent(m)
	if err != nil {
		return err
	}
	return builder.SetEvent(e)
}

func (b *binaryBufferedMessage) Finish(error) error {
	if b.body != nil {
		binaryMessagePool.Put(b.body)
	}
	return nil
}

// Binary Encoder
func (b *binaryBufferedMessage) SetData(data io.Reader) error {
	buf := binaryMessagePool.Get()
	w, err := io.Copy(buf, data)
	if err != nil {
		return err
	}
	if w == 0 {
		binaryMessagePool.Put(buf)
		return nil
	}
	b.body = buf
	return nil
}

func (b *binaryBufferedMessage) SetAttribute(attribute spec.Attribute, value interface{}) error {
	// If spec version we need to change to right context struct
	b.metadata[attribute] = value
	return nil
}

func (b *binaryBufferedMessage) SetExtension(name string, value interface{}) error {
	b.extensions[name] = value
	return nil
}

var _ binding.Message = (*binaryBufferedMessage)(nil) // Test it conforms to the interface
var _ binding.BinaryEncoder = (*binaryBufferedMessage)(nil)
