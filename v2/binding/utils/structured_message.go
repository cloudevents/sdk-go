package utils

import (
	"context"
	"io"

	"github.com/cloudevents/sdk-go/v2/binding"
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

var _ binding.Message = (*genericStructuredMessage)(nil)

func (m *genericStructuredMessage) ReadEncoding() binding.Encoding {
	return binding.EncodingStructured
}

func (m *genericStructuredMessage) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	return encoder.SetStructuredEvent(ctx, m.format, m.reader)
}

func (m *genericStructuredMessage) ReadBinary(ctx context.Context, encoder binding.BinaryWriter) error {
	return binding.ErrNotBinary
}

func (m *genericStructuredMessage) Finish(err error) error {
	if closer, ok := m.reader.(io.ReadCloser); ok {
		if err2 := closer.Close(); err2 != nil {
			return err2
		}
	}
	return err
}
