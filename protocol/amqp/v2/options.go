/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"github.com/Azure/go-amqp"
)

// Option is the function signature required to be considered an amqp.Option.
// Options are applied to the Protocol during construction to configure
// sender and receiver link behavior.
type Option func(*Protocol) error

// WithSenderOptions sets sender options for the AMQP sender link.
// If called multiple times, later calls will override earlier ones.
func WithSenderOptions(opts *amqp.SenderOptions) Option {
	return func(t *Protocol) error {
		t.senderLinkOpts = opts
		return nil
	}
}

// WithReceiverOptions sets receiver options for the AMQP receiver link.
// If called multiple times, later calls will override earlier ones.
func WithReceiverOptions(opts *amqp.ReceiverOptions) Option {
	return func(t *Protocol) error {
		t.receiverLinkOpts = opts
		return nil
	}
}

// SenderOptionFunc is the type of amqp.Sender options
type SenderOptionFunc func(sender *sender)
