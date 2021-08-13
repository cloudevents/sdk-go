/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package format

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/cloudevents/sdk-go/v2/event/datacodec"
)

const (
	// ContentTypeProtobuf indicates that the data attribute is a protobuf
	// message.
	ContentTypeProtobuf = "application/protobuf"
)

func init() {
	datacodec.AddDecoder(ContentTypeProtobuf, DecodeData)
	datacodec.AddEncoder(ContentTypeProtobuf, EncodeData)
}

// DecodeData converts an encoded protobuf message back into the message (out).
// The message must be a type compatible with whatever was given to EncodeData.
func DecodeData(ctx context.Context, in []byte, out interface{}) error {
	outmsg, ok := out.(proto.Message)
	if !ok {
		return fmt.Errorf("can only decode protobuf into proto.Message. got %T", out)
	}
	if err := proto.Unmarshal(in, outmsg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %s", err)
	}
	return nil
}

// EncodeData a protobuf message to bytes.
//
// Like the official datacodec implementations, this one returns the given value
// as-is if it is already a byte slice.
func EncodeData(ctx context.Context, in interface{}) ([]byte, error) {
	if b, ok := in.([]byte); ok {
		return b, nil
	}
	var pbmsg proto.Message
	var ok bool
	if pbmsg, ok = in.(proto.Message); !ok {
		return nil, fmt.Errorf("protobuf encoding only works with protobuf messages. got %T", in)
	}
	return proto.Marshal(pbmsg)
}
