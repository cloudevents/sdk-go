package buffering

import (
	"bytes"
	"io"

	"github.com/valyala/bytebufferpool"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/event"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

var specs = spec.New()
var binaryMessagePool bytebufferpool.Pool

// binaryBufferedMessage implements a binary-mode message as a simple struct.
// This message implementation is used by CopyMessage and BufferMessage
type binaryBufferedMessage struct {
	context ce.EventContext
	body    *bytebufferpool.ByteBuffer
}

func (m *binaryBufferedMessage) Structured(binding.StructuredEncoder) error {
	return binding.ErrNotStructured
}

func (m *binaryBufferedMessage) Binary(b binding.BinaryEncoder) (err error) {
	err = event.EventContextToBinaryEncoder(m.context, b)
	if err != nil {
		return err
	}
	if m.body != nil {
		return b.SetData(bytes.NewReader(m.body.Bytes()))
	}
	return err
}

func (m *binaryBufferedMessage) Event(builder binding.EventEncoder) error {
	// We must copy the body to don't cause memory leaks
	if m.body != nil {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, bytes.NewReader(m.body.Bytes()))
		if err != nil {
			return err
		}

		return builder.SetEvent(ce.Event{
			Context:     m.context,
			Data:        buf.Bytes(),
			DataEncoded: true,
			DataBinary:  true,
		})
	}
	return builder.SetEvent(ce.Event{
		Context: m.context,
	})
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
	if attribute.Kind() == spec.SpecVersion {
		str, err := types.ToString(value)
		if err != nil {
			return err
		}
		switch str {
		case cloudevents.VersionV01:
			b.context = b.context.AsV01()
		case cloudevents.VersionV02:
			b.context = b.context.AsV02()
		case cloudevents.VersionV03:
			b.context = b.context.AsV03()
		case cloudevents.VersionV1:
			b.context = b.context.AsV1()
		}
		return nil
	}
	return attribute.Set(b.context, value)
}

func (b *binaryBufferedMessage) SetExtension(name string, value interface{}) error {
	value, err := types.Validate(value)
	if err != nil {
		return err
	}
	return b.context.SetExtension(name, value)
}

var _ binding.Message = (*binaryBufferedMessage)(nil) // Test it conforms to the interface
var _ binding.BinaryEncoder = (*binaryBufferedMessage)(nil)
