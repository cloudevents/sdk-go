package binding

import (
	"context"
	"io"

	"github.com/cloudevents/sdk-go/v2/binding/format"
)

type genericStructuredMessage struct {
	format format.Format
	reader io.Reader
}

// NewStructuredMessage wraps a format and an io.Reader returning an implementation of Message
// This message *cannot* be read several times safely
func NewStructuredMessage(format format.Format, reader io.Reader) *genericStructuredMessage {
	return &genericStructuredMessage{reader: reader, format: format}
}

var _ Message = (*genericStructuredMessage)(nil)

func (m *genericStructuredMessage) ReadEncoding() Encoding {
	return EncodingStructured
}

func (m *genericStructuredMessage) ReadStructured(ctx context.Context, encoder StructuredWriter) error {
	return encoder.SetStructuredEvent(ctx, m.format, m.reader)
}

func (m *genericStructuredMessage) ReadBinary(ctx context.Context, encoder BinaryWriter) error {
	return ErrNotBinary
}

func (m *genericStructuredMessage) Finish(err error) error {
	return nil
}
