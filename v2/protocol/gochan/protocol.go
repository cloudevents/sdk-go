/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package gochan

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

const (
	defaultChanDepth = 20
)

// SendReceiver is a reference implementation for using the CloudEvents binding
// integration.
type SendReceiver struct {
	sender   protocol.SendCloser
	receiver protocol.Receiver
}

func New() *SendReceiver {
	ch := make(chan binding.Message, defaultChanDepth)

	return &SendReceiver{
		sender:   Sender(ch),
		receiver: Receiver(ch),
	}
}

func (sr *SendReceiver) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) (err error) {
	return sr.sender.Send(ctx, in, transformers...)
}

func (sr *SendReceiver) Receive(ctx context.Context) (binding.Message, error) {
	return sr.receiver.Receive(ctx)
}

func (sr *SendReceiver) Close(ctx context.Context) error {
	return sr.sender.Close(ctx)
}
