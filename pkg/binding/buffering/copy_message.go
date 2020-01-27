package buffering

import (
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/event"
)

// BufferMessage does the same than CopyMessage and it also bounds the original Message
// lifecycle to the newly created message: calling Finish() on the returned message calls m.Finish()
func BufferMessage(m binding.Message) (binding.Message, error) {
	result, err := CopyMessage(m)
	if err != nil {
		return nil, err
	}
	return binding.WithFinish(result, func(err error) { _ = m.Finish(err) }), nil
}

// CopyMessage reads m once and creates an in-memory copy depending on the encoding of m.
// The returned copy is not dependent on any transport and can be read many times.
// When the copy can be forgot, the copied message must be finished with Finish() message to release the memory
func CopyMessage(m binding.Message) (binding.Message, error) {
	// Try structured first, it's cheaper.
	sm := structBufferedMessage{}
	err := m.Structured(&sm)
	switch err {
	case nil:
		return &sm, nil
	case binding.ErrNotStructured:
		break
	default:
		return nil, err
	}
	bm := binaryBufferedMessage{context: &cloudevents.EventContextV1{}}
	err = m.Binary(&bm)
	switch err {
	case nil:
		return &bm, nil
	case binding.ErrNotBinary:
		break
	default:
		return nil, err
	}

	em := event.EventMessage{}
	err = m.Event(&em)
	if err != nil {
		return nil, err
	}
	return &em, nil
}
