/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package stan

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/binding"
)

// MsgMetadata holds metadata of a received *stan.Msg. This information is kept in a separate struct so that users
// cannot interact with the underlying message outside of the SDK.
type MsgMetadata struct {
	Sequence        uint64
	Redelivered     bool
	RedeliveryCount uint32
}

type msgKeyType struct{}

var msgKey msgKeyType

// MetadataContextDecorator returns an inbound context decorator which adds STAN message metadata to
// the current context. If the inbound message is not a *stan.Message then this decorator is a no-op.
func MetadataContextDecorator() func(context.Context, binding.Message) context.Context {
	return func(ctx context.Context, m binding.Message) context.Context {
		if msg, ok := m.(*Message); ok {
			return context.WithValue(ctx, msgKey, MsgMetadata{
				Sequence:        msg.Msg.Sequence,
				Redelivered:     msg.Msg.Redelivered,
				RedeliveryCount: msg.Msg.RedeliveryCount,
			})
		}

		return ctx
	}
}

// MessageMetadataFrom extracts the STAN message metadata from the provided ctx. The bool return parameter is true if
// the metadata was set on the context, or false otherwise.
func MessageMetadataFrom(ctx context.Context) (MsgMetadata, bool) {
	if v, ok := ctx.Value(msgKey).(MsgMetadata); ok {
		return v, true
	}

	return MsgMetadata{}, false
}
