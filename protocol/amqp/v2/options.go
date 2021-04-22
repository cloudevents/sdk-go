/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package amqp

import (
	"github.com/Azure/go-amqp"
)

// Option is the function signature required to be considered an amqp.Option.
type Option func(*Protocol) error

// WithConnOpt sets a connection option for amqp
func WithConnOpt(opt amqp.ConnOption) Option {
	return func(t *Protocol) error {
		t.connOpts = append(t.connOpts, opt)
		return nil
	}
}

// WithConnSASLPlain sets SASLPlain connection option for amqp
func WithConnSASLPlain(username, password string) Option {
	return WithConnOpt(amqp.ConnSASLPlain(username, password))
}

// WithSessionOpt sets a session option for amqp
func WithSessionOpt(opt amqp.SessionOption) Option {
	return func(t *Protocol) error {
		t.sessionOpts = append(t.sessionOpts, opt)
		return nil
	}
}

// WithSenderLinkOption sets a link option for amqp
func WithSenderLinkOption(opt amqp.LinkOption) Option {
	return func(t *Protocol) error {
		t.senderLinkOpts = append(t.senderLinkOpts, opt)
		return nil
	}
}

// WithReceiverLinkOption sets a link option for amqp
func WithReceiverLinkOption(opt amqp.LinkOption) Option {
	return func(t *Protocol) error {
		t.receiverLinkOpts = append(t.receiverLinkOpts, opt)
		return nil
	}
}

// SenderOptionFunc is the type of amqp.Sender options
type SenderOptionFunc func(sender *sender)
