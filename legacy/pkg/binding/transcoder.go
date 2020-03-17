package binding

import ce "github.com/cloudevents/sdk-go/legacy"

// Implements a transformation process while transferring the event from the Message implementation
// to the provided encoder
//
// A transformer could optionally not provide an implementation for binary and/or structured encodings,
// returning nil to the respective factory method.
type TransformerFactory interface {
	// Can return nil if the transformation doesn't support structured encoding directly
	StructuredTransformer(encoder StructuredEncoder) StructuredEncoder

	// Can return nil if the transformation doesn't support binary encoding directly
	BinaryTransformer(encoder BinaryEncoder) BinaryEncoder

	// Can return nil if the transformation doesn't support events
	EventTransformer() EventTransformer
}

// Utility type alias to manage multiple TransformerFactory
type TransformerFactories []TransformerFactory

func (t TransformerFactories) StructuredTransformer(encoder StructuredEncoder) StructuredEncoder {
	if encoder == nil {
		return nil
	}
	res := encoder
	for _, b := range t {
		if new := b.StructuredTransformer(res); new != nil {
			res = new
		} else {
			return nil // Structured not supported!
		}
	}
	return res
}

func (t TransformerFactories) BinaryTransformer(encoder BinaryEncoder) BinaryEncoder {
	if encoder == nil {
		return nil
	}
	res := encoder
	for _, b := range t {
		if new := b.BinaryTransformer(res); new != nil {
			res = new
		} else {
			return nil // Binary not supported!
		}
	}
	return res
}

func (t TransformerFactories) EventTransformer() EventTransformer {
	return func(e *ce.Event) error {
		for _, factory := range t {
			f := factory.EventTransformer()

			if f != nil {
				err := f(e)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

type EventTransformer func(*ce.Event) error
