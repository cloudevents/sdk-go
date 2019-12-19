package binding

// Invokes the encoders. createRootStructuredEncoder and createRootBinaryEncoder could be null if the protocol doesn't support it
//
// Returns:
// * true, false, nil if message was structured and correctly translated to Event
// * false, true, nil if message was binary and correctly translated to Event
// * false, false, nil if message was event and correctly translated to Event
// * true, false, err if message was structured but error happened during translation
// * false, true, err if message was binary but error happened during translation
// * false, false, err if message was event but error happened during translation
// * false, false, err in other cases
func Translate(
	message Message,
	createRootStructuredEncoder func() StructuredEncoder,
	createRootBinaryEncoder func() BinaryEncoder,
	createRootEventEncoder func() EventEncoder,
	factories TransformerFactories,
) (bool, bool, error) {
		if createRootStructuredEncoder != nil {
			// Wrap the transformers in the structured builder
			structuredEncoder := factories.StructuredTransformer(createRootStructuredEncoder())

			// StructuredTransformer could return nil if one of transcoders doesn't support
			// direct structured transcoding
			if structuredEncoder != nil {
				if err := message.Structured(structuredEncoder); err == nil {
					return true, false, nil
				} else if err != ErrNotStructured {
					return true, false, err
				}
			}
		}

		if createRootBinaryEncoder != nil {
			binaryEncoder := factories.BinaryTransformer(createRootBinaryEncoder())
			if binaryEncoder != nil {
				if err := message.Binary(binaryEncoder); err == nil {
					return false, true, nil
				} else if err != ErrNotBinary {
					return false, true, err
				}
			}
		}

	eventEncoder := factories.EventTransformer(createRootEventEncoder())
	return false, false, message.Event(eventEncoder)
}
