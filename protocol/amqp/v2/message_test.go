/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"context"
	"testing"

	"github.com/Azure/go-amqp"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	. "github.com/cloudevents/sdk-go/v2/test"
)

func TestNewMessage_success(t *testing.T) {
	tests := []struct {
		name     string
		encoding binding.Encoding
	}{
		{
			name:     "Structured encoding",
			encoding: binding.EncodingStructured,
		},
		{
			name:     "Binary encoding",
			encoding: binding.EncodingBinary,
		},
	}
	for _, tt := range tests {
		EachEvent(t, Events(), func(t *testing.T, eventIn event.Event) {
			t.Run(tt.name, func(t *testing.T) {
				eventIn = eventIn.Clone()

				ctx := context.TODO()
				if tt.encoding == binding.EncodingStructured {
					ctx = binding.WithForceStructured(ctx)
				} else if tt.encoding == binding.EncodingBinary {
					ctx = binding.WithForceBinary(ctx)
				}

				message := amqp.Message{}
				require.NoError(t, WriteMessage(ctx, binding.ToMessage(&eventIn), &message))

				rcv := amqp.Receiver{}

				got := NewMessage(&message, &rcv)
				require.Equal(t, tt.encoding, got.ReadEncoding())
			})
		})
	}
}

func TestNewMessage_message_unknown(t *testing.T) {
	message := amqp.NewMessage([]byte("hello-world"))
	rcv := amqp.Receiver{}

	got := NewMessage(message, &rcv)
	require.Equal(t, binding.EncodingUnknown, got.ReadEncoding())
}
