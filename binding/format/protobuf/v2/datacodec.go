package format

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	// ContentTypeProtobuf indicates that the data attribute is a protobuf
	// message.
	ContentTypeProtobuf = "application/protobuf"
)

// DecodeData converts an encoded protobuf message back into the message (out).
// The message must be a type compatible with whatever was given to EncodeData.
// This method will assume that the encoded values was encoded by this package
// which wraps the content in an Any type but will fall back to directly
// unmarshalling the bytes into the message if that fails.
func DecodeData(ctx context.Context, in []byte, out interface{}) error {
	outmsg, ok := out.(proto.Message)
	if !ok {
		return fmt.Errorf("can only decode protobuf into proto.Message. got %T", out)
	}
	anymsg := &anypb.Any{}
	err := proto.Unmarshal(in, anymsg)
	if err != nil {
		if nextErr := proto.Unmarshal(in, outmsg); nextErr != nil {
			finalErr := fmt.Errorf("failed to directly unmarshal message: %s", nextErr)
			return fmt.Errorf("failed to unmarshal Any: %s -> %s", err, finalErr)
		}
	}
	// Intentionally not using anypb.UnmarshalTo because it compares the type
	// string from the Any value with the self-reported type string of the
	// destination message. This adds an artificial requirement to have any
	// local code generation be configured identical to the system that sent
	// the message even if the content is compatible. Instead, we'll attempt
	// to unmarshal into any compatible type to maximize the loose coupling.
	return proto.Unmarshal(anymsg.Value, outmsg)
}

// EncodeData a protobuf message to bytes. This implementations wraps the given
// message in an Any type before marshaling. This is done because using the Any
// type is a requirement when using the protobuf message format but the type
// information required to build an Any is only available here, before encoding.
// Wrapping the message in an Any allows the protobuf format to detect the
// encoding and unmarshal the bytes without knowing the underlying type
// information. The DecodeData method provided by this package does the inverse
// so that most users of the SDK are unaware that this happens.
//
// Like the official datacodec implementations, this one returns the given value
// as-is if it is already a byte slice. Additionally, if the value is an any
// message already then we preserve it as is since this is the equivalent case
// for protobuf where the value is already encoded.
func EncodeData(ctx context.Context, in interface{}) ([]byte, error) {
	if b, ok := in.([]byte); ok {
		return b, nil
	}
	var pbmsg proto.Message
	var ok bool
	if pbmsg, ok = in.(proto.Message); !ok {
		return nil, fmt.Errorf("protobuf encoding only works with protobuf messages. got %T", in)
	}
	anymsg, err := anypb.New(pbmsg)
	if err != nil {
		return nil, err
	}
	if v, ok := pbmsg.(*anypb.Any); ok {
		anymsg = v
	}
	// NOTE: We have to return the byte marshaled version here because the SDK
	// only has support for carrying bytes through the Event abstraction. This
	// is because bytes are the compatible format for the data field of all
	// message formats. The protobuf message format will have to inspect the
	// content type of the data, unmarshal protobuf messages back to an any
	// type, and then marshal the event with the any attached.
	return proto.Marshal(anymsg)
}
