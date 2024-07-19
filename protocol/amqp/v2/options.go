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
func WithConnOpts(opt *amqp.ConnOptions) Option {
	return func(t *Protocol) error {
		t.connOpts = opt
		return nil
	}
}

// WithConnSASLPlain sets SASLPlain connection option for amqp
func WithConnSASLPlain(opt *amqp.ConnOptions, username, password string) Option {
	opt.SASLType = amqp.SASLTypePlain(username, password)
	return WithConnOpts(opt)
}

// WithSessionOpt sets a session option for amqp
func WithSessionOpts(opt *amqp.SessionOptions) Option {
	return func(t *Protocol) error {
		t.sessionOpts = opt
		return nil
	}
}

// WithSenderLinkOption sets a link option for amqp
func WithSenderOpts(opt *amqp.SenderOptions) Option {
	return func(t *Protocol) error {
		t.senderOpts = opt
		return nil
	}
}

// WithReceiverLinkOption sets a link option for amqp
func WithReceiverOpts(opt *amqp.ReceiverOptions) Option {
	return func(t *Protocol) error {
		t.receiverOpts = opt
		return nil
	}
}

func WithReceiveOpts(opt amqp.ReceiveOptions) Option {
	return func(t *Protocol) error {
		t.receiveOpts = opt
		return nil
	}
}

func WithSendOpts(opt *amqp.SendOptions) Option {
	return func(t *Protocol) error {
		t.sendOpts = opt
		return nil
	}
}

// SenderOptionFunc is the type of amqp.Sender options
type SenderOptionFunc func(sender *sender)
