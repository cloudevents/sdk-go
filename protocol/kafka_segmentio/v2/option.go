/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_segmentio

import (
	"context"
)

// SenderOptionFunc is the type of kafka_segmentio.Sender options
type SenderOptionFunc func(sender *Sender)

// ProtocolOptionFunc is the type of kafka_segmentio.Protocol options
type ProtocolOptionFunc func(protocol *Protocol)

func WithReceiverGroupID(groupId string) ProtocolOptionFunc {
	return func(protocol *Protocol) {
		protocol.receiverGroupID = groupId
	}
}

func WithSenderContextDecorators(decorator func(context.Context) context.Context) ProtocolOptionFunc {
	return func(protocol *Protocol) {
		protocol.SenderContextDecorators = append(protocol.SenderContextDecorators, decorator)
	}
}
