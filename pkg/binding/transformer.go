package binding

import (
	"github.com/cloudevents/sdk-go/pkg/event"
)

// Implements a transformation process while transferring the event from the Message implementation
// to the provided encoder
//
// A transformer could optionally not provide an implementation for binary and/or structured encodings,
// returning nil to the respective factory method.
type TransformerFactory interface {
	// Can return nil if the transformation doesn't support structured encoding directly
	StructuredTransformer(encoder StructuredWriter) StructuredWriter

	// Can return nil if the transformation doesn't support binary encoding directly
	BinaryTransformer(encoder BinaryWriter) BinaryWriter

	// Can return nil if the transformation doesn't support events
	EventTransformer() EventTransformer
}

// Utility type alias to manage multiple TransformerFactory
type TransformerFactories []TransformerFactory

func (t TransformerFactories) StructuredTransformer(encoder StructuredWriter) StructuredWriter {
	if encoder == nil {
		return nil
	}
	res := encoder
	if t != nil {
		for _, b := range t {
			if new := b.StructuredTransformer(res); new != nil {
				res = new
			} else {
				return nil // Structured not supported!
			}
		}
	}
	return res
}

func (t TransformerFactories) BinaryTransformer(encoder BinaryWriter) BinaryWriter {
	if encoder == nil {
		return nil
	}
	res := encoder
	if t != nil {
		for i, _ := range t {
			if new := t[len(t)-i-1].BinaryTransformer(res); new != nil {
				res = new
			} else {
				return nil // Binary not supported!
			}
		}
	}
	return res
}

func (t TransformerFactories) EventTransformer() EventTransformer {
	return func(e *event.Event) error {
		if t != nil {
			for _, factory := range t {
				f := factory.EventTransformer()

				if f != nil {
					err := f(e)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
}

// EventTransformer mutates the provided Event
type EventTransformer func(*event.Event) error
