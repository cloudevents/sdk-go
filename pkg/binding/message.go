package binding

import (
	"github.com/cloudevents/sdk-go/pkg/binding/format"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

// EventMessage type-converts a cloudevents.Event object to implement Message.
// This allows local cloudevents.Event objects to be sent directly via Sender.Send()
//     s.Send(ctx, binding.EventMessage(e))
type EventMessage ce.Event

func (m EventMessage) Event() (ce.Event, error)     { return ce.Event(m), nil }
func (m EventMessage) Structured() (string, []byte) { return "", nil }
func (EventMessage) Finish(error) error             { return nil }

var _ Message = EventMessage{} // Test it conforms to the interface

// StructMessage implements a structured-mode message as a simple struct.
type StructMessage struct {
	Format string
	Bytes  []byte
}

// Event can only decode built-in formats supported by format.Unmarshal.
func (m StructMessage) Event() (e ce.Event, err error) {
	err = format.Unmarshal(m.Format, m.Bytes, &e)
	return e, err
}
func (m StructMessage) Structured() (string, []byte) { return m.Format, m.Bytes }
func (StructMessage) Finish(error) error             { return nil }

var _ Message = StructMessage{} // Test it conforms to the interface

// StructEncoder encodes events as StructMessage using a Format.
type StructEncoder struct{ Format format.Format }

func (enc StructEncoder) Encode(e ce.Event) (Message, error) {
	b, err := enc.Format.Marshal(e)
	return StructMessage{Format: enc.Format.MediaType(), Bytes: b}, err
}

type BinaryEncoder struct{}

func (BinaryEncoder) Encode(e ce.Event) (Message, error) { return EventMessage(e), nil }

// Structured returns the structured encoding of a message using a format.
// m.Structured() returns the correct format, return that.
// Otherwise use format the message's event with f.Format().
func Structured(m Message, f format.Format) ([]byte, error) {
	mt, b := m.Structured()
	if mt == f.MediaType() {
		return b, nil
	}
	e, err := m.Event()
	if err != nil {
		return nil, err
	}
	return f.Marshal(e)
}

type finishMessage struct {
	Message
	finish func(error)
}

func (m finishMessage) Finish(err error) error {
	err2 := m.Message.Finish(err) // Finish original message first
	if m.finish != nil {
		m.finish(err) // Notify callback
	}
	return err2
}

// WithFinish returns a wrapper for m that calls finish() and
// m.Finish() in its Finish().
// Allows code to be notified when a message is Finished.
func WithFinish(m Message, finish func(error)) Message {
	return finishMessage{Message: m, finish: finish}
}

// Translate creates a new message with the same content and mode as 'in',
// using the given makeBinary or makeStruct functions.
func Translate(in Message,
	makeBinary func(ce.Event) (Message, error),
	makeStruct func(string, []byte) (Message, error),
) (Message, error) {
	if f, b := in.Structured(); f != "" && len(b) > 0 {
		return makeStruct(f, b)
	}
	e, err := in.Event()
	if err != nil {
		return nil, err
	}
	return makeBinary(e)
}
