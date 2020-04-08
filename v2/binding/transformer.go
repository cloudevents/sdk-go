package binding

// TransformerFactory is an interface that implements a transformation
// process while transferring the event from the Message
// implementation to the provided encoder
//
// When a write function (binding.Write, binding.ToEvent, buffering.CopyMessage, etc.)
// takes Transformer(s) as parameter, it eventually converts the message to a form
// which correctly implements MessageMetadataReader, in order to guarantee that transformation
// is applied
type Transformer interface {
	Transform(MessageMetadataReader, MessageMetadataWriter) error
}

// TODO doc
type TransformerFunc func(MessageMetadataReader, MessageMetadataWriter) error

func (t TransformerFunc) Transform(r MessageMetadataReader, w MessageMetadataWriter) error {
	return t(r, w)
}

var _ Transformer = (TransformerFunc)(nil)

// TODO doc
type Transformers []Transformer

func (t Transformers) Transform(r MessageMetadataReader, w MessageMetadataWriter) error {
	for _, transformer := range t {
		err := transformer.Transform(r, w)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ Transformer = (Transformers)(nil)
