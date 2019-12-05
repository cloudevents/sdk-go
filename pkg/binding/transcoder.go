package binding

// Implements a transcoding process while transferring the event from the Message implementation
// to one of the builders
type TranscoderFactory interface {
	// Can return nil if the transformation doesn't support structured encoding directly
	StructuredMessageTranscoder(builder StructuredMessageBuilder) StructuredMessageBuilder

	// Can return nil if the transformation doesn't support binary encoding directly
	BinaryMessageTranscoder(builder BinaryMessageBuilder) BinaryMessageBuilder

	// Cannot return nil
	EventMessageTranscoder(builder EventMessageBuilder) EventMessageBuilder
}

type TranscoderFactories []TranscoderFactory

func (t TranscoderFactories) StructuredMessageTranscoder(builder StructuredMessageBuilder) StructuredMessageBuilder {
	res := builder
	for _, b := range t {
		if new := b.StructuredMessageTranscoder(res); new != nil {
			res = new
		} else {
			return nil // Structured not supported!
		}
	}
	return res
}

func (t TranscoderFactories) BinaryMessageTranscoder(builder BinaryMessageBuilder) BinaryMessageBuilder {
	res := builder
	for _, b := range t {
		if new := b.BinaryMessageTranscoder(res); new != nil {
			res = new
		} else {
			return nil // Binary not supported!
		}
	}
	return res
}

func (t TranscoderFactories) EventMessageTranscoder(builder EventMessageBuilder) EventMessageBuilder {
	res := builder
	for _, b := range t {
		res = b.EventMessageTranscoder(res)
	}
	return res
}
