package binding

import (
	"context"

	"github.com/cloudevents/sdk-go/pkg/event"
)

const (
	SKIP_DIRECT_STRUCTURED_ENCODING = "SKIP_DIRECT_STRUCTURED_ENCODING"
	SKIP_DIRECT_BINARY_ENCODING     = "SKIP_DIRECT_BINARY_ENCODING"
	PREFERRED_EVENT_ENCODING        = "PREFERRED_EVENT_ENCODING"
)

// Invokes the encoders. structuredEncoder and binaryEncoder could be nil if the protocol doesn't support it.
// transformers can be nil and this function guarantees that they are invoked only once during the encoding process.
//
// Returns:
// * EncodingStructured, nil if message is correctly encoded in structured encoding
// * EncodingBinary, nil if message is correctly encoded in binary encoding
// * EncodingStructured, err if message was structured but error happened during the encoding
// * EncodingBinary, err if message was binary but error happened during the encoding
// * EncodingUnknown, ErrUnknownEncoding if message is not a structured or a binary Message
func RunDirectEncoding(
	ctx context.Context,
	message Message,
	structuredEncoder StructuredEncoder,
	binaryEncoder BinaryEncoder,
	transformers TransformerFactories,
) (Encoding, error) {
	if structuredEncoder != nil && !GetOrDefaultFromCtx(ctx, SKIP_DIRECT_STRUCTURED_ENCODING, false).(bool) {
		// Wrap the transformers in the structured builder
		structuredEncoder = transformers.StructuredTransformer(structuredEncoder)

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
		binaryEncoder = transformers.BinaryTransformer(binaryEncoder)
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
// 1. It first tries direct encoding using RunDirectEncoding
// 2. If no direct encoding is possible, it uses ToEvent to generate an Event representation
// 3. From the Event, the message is encoded back to the provided structured or binary encoders
// You can tweak the encoding process using the context decorators WithForceStructured, WithForceStructured, etc.
// transformers can be nil and this function guarantees that they are invoked only once during the encoding process.
// Returns:
// * EncodingStructured, nil if message is correctly encoded in structured encoding
// * EncodingBinary, nil if message is correctly encoded in binary encoding
// * EncodingUnknown, ErrUnknownEncoding if message.Encoding() == EncodingUnknown
// * _, err if error happened during the encoding
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

	var e event.Event
	e, err = ToEvent(ctx, message, transformers)
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

	return EncodingUnknown, ErrUnknownEncoding
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

// Get a configuration value from the provided context
func GetOrDefaultFromCtx(ctx context.Context, key string, def interface{}) interface{} {
	if val := ctx.Value(key); val != nil {
		return val
	} else {
		return def
	}
}
