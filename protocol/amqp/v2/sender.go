/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"context"

	"github.com/Azure/go-amqp"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

// sender wraps an amqp.Sender as a binding.Sender
type sender struct {
	amqp *amqp.Sender
}

func (s *sender) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) error {
	var err error
	defer func() { _ = in.Finish(err) }()
	if m, ok := in.(*Message); ok { // Already an AMQP message.
		err = s.amqp.Send(ctx, m.AMQP)
		return err
	}

	var amqpMessage amqp.Message
	err = WriteMessage(ctx, in, &amqpMessage, transformers...)
	if err != nil {
		return err
	}

	err = s.amqp.Send(ctx, &amqpMessage)
	return err
}

// NewSender creates a new Sender which wraps an amqp.Sender in a binding.Sender
func NewSender(amqpSender *amqp.Sender, options ...SenderOptionFunc) protocol.Sender {
	s := &sender{amqp: amqpSender}
	for _, o := range options {
		o(s)
	}
	return s
}
