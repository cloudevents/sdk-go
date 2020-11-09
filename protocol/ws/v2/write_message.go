package v2

import (
	"context"
	"io"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
)

// WriteWriter fills the provided writer with the bindings.Message m.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
func WriteWriter(ctx context.Context, m binding.Message, writer io.WriteCloser, transformers ...binding.Transformer) error {
	structuredWriter := &wsMessageWriter{writer}

	_, err := binding.Write(
		ctx,
		m,
		structuredWriter,
		nil,
		transformers...,
	)
	return err
}

type wsMessageWriter struct {
	io.WriteCloser
}

func (w *wsMessageWriter) SetStructuredEvent(_ context.Context, _ format.Format, event io.Reader) error {
	_, err := io.Copy(w.WriteCloser, event)
	if err != nil {
		// Try to close anyway
		_ = w.WriteCloser.Close()
		return err
	}

	return w.WriteCloser.Close()
}

var _ binding.StructuredWriter = (*wsMessageWriter)(nil) // Test it conforms to the interface
