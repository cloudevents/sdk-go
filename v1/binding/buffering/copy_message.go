package buffering

import (
	"context"

	"github.com/cloudevents/sdk-go/v1/binding"
)

// BufferMessage does the same than CopyMessage and it also bounds the original Message
// lifecycle to the newly created message: calling Finish() on the returned message calls m.Finish()
func BufferMessage(ctx context.Context, m binding.Message, transformers ...binding.TransformerFactory) (binding.Message, error) {
	result, err := CopyMessage(ctx, m, transformers...)
	if err != nil {
		return nil, err
	}
	return binding.WithFinish(result, func(err error) { _ = m.Finish(err) }), nil
}

// CopyMessage reads m once and creates an in-memory copy depending on the encoding of m.
// The returned copy is not dependent on any transport and can be read many times.
// When the copy can be forgot, the copied message must be finished with Finish() message to release the memory
func CopyMessage(ctx context.Context, m binding.Message, transformers ...binding.TransformerFactory) (binding.Message, error) {
	originalMessageEncoding := m.Encoding()

	if originalMessageEncoding == binding.EncodingUnknown {
		return nil, binding.ErrUnknownEncoding
	}
	if originalMessageEncoding == binding.EncodingEvent {
		e, _, err := binding.ToEvent(ctx, m, transformers...)
		if err != nil {
			return nil, err
		}
		return binding.EventMessage(e), nil
	}

	sm := structBufferedMessage{}
	bm := binaryBufferedMessage{}

	encoding, err := binding.RunDirectEncoding(ctx, m, &sm, &bm, transformers)
	if encoding == binding.EncodingStructured {
		return &sm, err
	} else if encoding == binding.EncodingBinary {
		return &bm, err
	} else {
		e, _, err := binding.ToEvent(ctx, m, transformers...)
		if err != nil {
			return nil, err
		}
		return binding.EventMessage(e), nil
	}
}
