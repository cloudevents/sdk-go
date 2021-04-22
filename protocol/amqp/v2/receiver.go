/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"context"
	"io"

	"github.com/Azure/go-amqp"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

// receiver wraps an amqp.Receiver as a binding.Receiver
type receiver struct{ amqp *amqp.Receiver }

func (r *receiver) Receive(ctx context.Context) (binding.Message, error) {
	m, err := r.amqp.Receive(ctx)
	if err != nil {
		if err == ctx.Err() {
			return nil, io.EOF
		}
		return nil, err
	}

	return NewMessage(m), nil
}

// NewReceiver create a new Receiver which wraps an amqp.Receiver in a binding.Receiver
func NewReceiver(amqp *amqp.Receiver) protocol.Receiver {
	return &receiver{amqp: amqp}
}
