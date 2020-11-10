package utils

import (
	"context"
	"io"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
)

// WriteStructured fills the provided io.Writer with the binding.Message m in structured mode.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
func WriteStructured(ctx context.Context, m binding.Message, writer io.Writer, transformers ...binding.Transformer) error {
	_, err := binding.Write(
		ctx,
		m,
		wsMessageWriter{writer},
		nil,
		transformers...,
	)
	return err
}

type wsMessageWriter struct {
	io.Writer
}

func (w wsMessageWriter) SetStructuredEvent(_ context.Context, _ format.Format, event io.Reader) error {
	_, err := io.Copy(w.Writer, event)
	if err != nil {
		// Try to close anyway
		_ = w.tryToClose()
		return err
	}

	return w.tryToClose()
}

func (w wsMessageWriter) tryToClose() error {
	if closer, ok := w.Writer.(io.WriteCloser); ok {
		return closer.Close()
	}
	return nil
}

var _ binding.StructuredWriter = wsMessageWriter{} // Test it conforms to the interface
