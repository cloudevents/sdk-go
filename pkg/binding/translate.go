package binding

import (
	"errors"

	ce "github.com/cloudevents/sdk-go"
)

// Invokes the encoders. createRootStructuredEncoder and createRootBinaryEncoder could be null if the protocol doesn't support it
//
// Returns:
// * EncodingStructured, nil if message was structured and correctly translated to Event
// * EncodingBinary, nil if message was binary and correctly translated to Event
// * EncodingStructured, err if message was structured but error happened during translation
// * BinaryEncoding, err if message was binary but error happened during translation
// * EncodingUnknown, ErrUnknownEncoding if message is not recognized
func RunDirectEncoding(
	message Message,
	structuredEncoder StructuredEncoder,
	binaryEncoder BinaryEncoder,
	factories TransformerFactories,
) (Encoding, error) {
	if structuredEncoder != nil {
		// Wrap the transformers in the structured builder
		structuredEncoder = factories.StructuredTransformer(structuredEncoder)

		// StructuredTransformer could return nil if one of transcoders doesn't support
		// direct structured transcoding
		if structuredEncoder != nil {
			if err := message.Structured(structuredEncoder); err == nil {
				return EncodingStructured, nil
			} else if err != ErrNotStructured {
				return EncodingStructured, err
			}
		}
	}

	if binaryEncoder != nil {
		binaryEncoder = factories.BinaryTransformer(binaryEncoder)
		if binaryEncoder != nil {
			if err := message.Binary(binaryEncoder); err == nil {
				return EncodingBinary, nil
			} else if err != ErrNotBinary {
				return EncodingBinary, err
			}
		}
	}

	return EncodingUnknown, ErrUnknownEncoding
}

// This is the full algorithm to encode a Message using transformers:
// 1. It first tries direct encoding using RunEncoders
// 2. If no direct encoding is possible, it goes through ToEvent to generate an event representation
// 3. Using the encoders previously defined
// Returns:
// * EncodingStructured, nil if message was structured and correctly translated to Event
// * EncodingBinary, nil if message was binary and correctly translated to Event
// * EncodingStructured, err if message was structured but error happened during translation
// * BinaryEncoding, err if message was binary but error happened during translation
// * EncodingUnknown, ErrUnknownEncoding if message is not recognized
func Encode(
	message Message,
	structuredEncoder StructuredEncoder,
	binaryEncoder BinaryEncoder,
	transformers TransformerFactories,
	eventPreferredEncoding Encoding,
) (Encoding, error) {
	enc := message.Encoding()
	var err error
	// Skip direct encoding if the event is an event message
	if enc != EncodingEvent {
		enc, err = RunDirectEncoding(message, structuredEncoder, binaryEncoder, transformers)
		if enc != EncodingUnknown {
			// Message directly encoded, nothing else to do here
			return enc, err
		}
	}

	var e ce.Event
	e, enc, err = ToEvent(message, transformers)
	if err != nil {
		return enc, err
	}

	message = EventMessage(e)

	if eventPreferredEncoding == EncodingStructured {
		if structuredEncoder != nil {
			return EncodingStructured, message.Structured(structuredEncoder)
		}
		if binaryEncoder != nil {
			return EncodingStructured, message.Binary(binaryEncoder)
		}
	} else {
		if binaryEncoder != nil {
			return EncodingStructured, message.Binary(binaryEncoder)
		}
		if structuredEncoder != nil {
			return EncodingStructured, message.Structured(structuredEncoder)
		}
	}

	return enc, errors.New("cannot find a suitable encoder to use from EventMessage")
}
