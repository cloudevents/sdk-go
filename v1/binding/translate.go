package binding

import (
	"context"
	"errors"

	ce "github.com/cloudevents/sdk-go/v1"
)

const (
	SKIP_DIRECT_STRUCTURED_ENCODING = "SKIP_DIRECT_STRUCTURED_ENCODING"
	SKIP_DIRECT_BINARY_ENCODING     = "SKIP_DIRECT_BINARY_ENCODING"
	PREFERRED_EVENT_ENCODING        = "PREFERRED_EVENT_ENCODING"
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
	ctx context.Context,
	message Message,
	structuredEncoder StructuredEncoder,
	binaryEncoder BinaryEncoder,
	factories TransformerFactories,
) (Encoding, error) {
	if structuredEncoder != nil && !GetOrDefaultFromCtx(ctx, SKIP_DIRECT_STRUCTURED_ENCODING, false).(bool) {
		// Wrap the transformers in the structured builder
		structuredEncoder = factories.StructuredTransformer(structuredEncoder)

		// StructuredTransformer could return nil if one of transcoders doesn't support
		// direct structured transcoding
		if structuredEncoder != nil {
			if err := message.Structured(ctx, structuredEncoder); err == nil {
				return EncodingStructured, nil
			} else if err != ErrNotStructured {
				return EncodingStructured, err
			}
		}
	}

	if binaryEncoder != nil && !GetOrDefaultFromCtx(ctx, SKIP_DIRECT_BINARY_ENCODING, false).(bool) {
		binaryEncoder = factories.BinaryTransformer(binaryEncoder)
		if binaryEncoder != nil {
			if err := message.Binary(ctx, binaryEncoder); err == nil {
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
// You can tweak the encoding process using the context decorators WithForceStructured, WithForceStructured, etc.
// This function guarantees that transformers are invoked only one time during the encoding process.
// Returns:
// * EncodingStructured, nil if message was structured and correctly translated to Event
// * EncodingBinary, nil if message was binary and correctly translated to Event
// * EncodingStructured, err if message was structured but error happened during translation
// * BinaryEncoding, err if message was binary but error happened during translation
// * EncodingUnknown, ErrUnknownEncoding if message is not recognized
func Encode(
	ctx context.Context,
	message Message,
	structuredEncoder StructuredEncoder,
	binaryEncoder BinaryEncoder,
	transformers TransformerFactories,
) (Encoding, error) {
	enc := message.Encoding()
	var err error
	// Skip direct encoding if the event is an event message
	if enc != EncodingEvent {
		enc, err = RunDirectEncoding(ctx, message, structuredEncoder, binaryEncoder, transformers)
		if enc != EncodingUnknown {
			// Message directly encoded, nothing else to do here
			return enc, err
		}
	}

	var e ce.Event
	e, enc, err = ToEvent(ctx, message, transformers)
	if err != nil {
		return enc, err
	}

	message = EventMessage(e)

	if GetOrDefaultFromCtx(ctx, PREFERRED_EVENT_ENCODING, EncodingBinary).(Encoding) == EncodingStructured {
		if structuredEncoder != nil {
			return EncodingStructured, message.Structured(ctx, structuredEncoder)
		}
		if binaryEncoder != nil {
			return EncodingBinary, message.Binary(ctx, binaryEncoder)
		}
	} else {
		if binaryEncoder != nil {
			return EncodingBinary, message.Binary(ctx, binaryEncoder)
		}
		if structuredEncoder != nil {
			return EncodingStructured, message.Structured(ctx, structuredEncoder)
		}
	}

	return enc, errors.New("cannot find a suitable encoder to use from EventMessage")
}

// Skip direct structured to structured encoding during the encoding process
func WithSkipDirectStructuredEncoding(ctx context.Context, skip bool) context.Context {
	return context.WithValue(ctx, SKIP_DIRECT_STRUCTURED_ENCODING, skip)
}

// Skip direct binary to binary encoding during the encoding process
func WithSkipDirectBinaryEncoding(ctx context.Context, skip bool) context.Context {
	return context.WithValue(ctx, SKIP_DIRECT_BINARY_ENCODING, skip)
}

// Define the preferred encoding from event to message during the encoding process
func WithPreferredEventEncoding(ctx context.Context, enc Encoding) context.Context {
	return context.WithValue(ctx, PREFERRED_EVENT_ENCODING, enc)
}

// Force structured encoding during the encoding process
func WithForceStructured(ctx context.Context) context.Context {
	return context.WithValue(context.WithValue(ctx, PREFERRED_EVENT_ENCODING, EncodingStructured), SKIP_DIRECT_BINARY_ENCODING, true)
}

// Force binary encoding during the encoding process
func WithForceBinary(ctx context.Context) context.Context {
	return context.WithValue(context.WithValue(ctx, PREFERRED_EVENT_ENCODING, EncodingBinary), SKIP_DIRECT_STRUCTURED_ENCODING, true)
}

func GetOrDefaultFromCtx(ctx context.Context, key string, def interface{}) interface{} {
	if val := ctx.Value(key); val != nil {
		return val
	} else {
		return def
	}
}
