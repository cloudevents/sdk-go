/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"errors"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var (
	ErrNoConnection                 = errors.New("URL or nats connection must be given")
	ErrConflictingConnection        = errors.New("URL and nats connection were both given")
	ErrNoFilterSubjects             = errors.New("no filter subjects were given")
	ErrMoreThanOneStream            = errors.New("more than one stream for given filter subjects")
	ErrNoConsumerConfig             = errors.New("no consumer config was given")
	ErrMoreThanOneConsumerConfig    = errors.New("more than one consumer config given")
	ErrNoSendSubject                = errors.New("no send subject given")
	ErrReceiverOptionsWithoutConfig = errors.New("receiver options given without consumer config")
	ErrSenderOptionsWithoutSubject  = errors.New("sender options given without send subject")
)

// ProtocolOption provides a way to configure the protocol
type ProtocolOption func(*Protocol) error

// WithURL creates a connection to be used in the protocol sender and receiver
// If WithConnection is also used, WithURL will be ignored.
func WithURL(url string, natsOpts ...nats.Option) ProtocolOption {
	return func(p *Protocol) error {
		opts := []nats.Option{}
		opts = append(opts, natsOpts...)
		p.url = url
		p.natsOpts = opts
		p.connOwned = true
		return nil
	}
}

// WithConnection uses the provided connection in the protocol sender and receiver
func WithConnection(conn *nats.Conn) ProtocolOption {
	return func(p *Protocol) error {
		p.conn = conn
		return nil
	}
}

// WithJetStreamOptions sets jetstream options used in the protocol sender and receiver
func WithJetStreamOptions(jetStreamOpts []jetstream.JetStreamOpt) ProtocolOption {
	return func(p *Protocol) error {
		p.jetSteamOpts = jetStreamOpts
		return nil
	}
}

// WithPublishOptions sets publish options used in the protocol sender
func WithPublishOptions(publishOpts []jetstream.PublishOpt) ProtocolOption {
	return func(p *Protocol) error {
		p.publishOpts = publishOpts
		return nil
	}
}

// WithSendSubject sets the subject used in the protocol sender
func WithSendSubject(sendSubject string) ProtocolOption {
	return func(p *Protocol) error {
		p.sendSubject = sendSubject
		return nil
	}
}

// WithConsumerConfig creates a unordered consumer used in the protocol receiver.
// This option is mutually exclusive with WithOrderedConsumerConfig.
func WithConsumerConfig(consumerConfig *jetstream.ConsumerConfig) ProtocolOption {
	return func(p *Protocol) error {
		if p.orderedConsumerConfig != nil {
			return ErrMoreThanOneConsumerConfig
		}
		p.consumerConfig = consumerConfig
		return nil
	}
}

// WithOrderedConsumerConfig creates a ordered consumer used in the protocol receiver.
// This option is mutually exclusive with WithConsumerConfig.
func WithOrderedConsumerConfig(orderedConsumerConfig *jetstream.OrderedConsumerConfig) ProtocolOption {
	return func(p *Protocol) error {
		if p.consumerConfig != nil {
			return ErrMoreThanOneConsumerConfig
		}
		p.orderedConsumerConfig = orderedConsumerConfig
		return nil
	}
}

// WithPullConsumerOptions sets pull options used in the protocol receiver.
func WithPullConsumerOptions(pullConsumeOpts []jetstream.PullConsumeOpt) ProtocolOption {
	return func(p *Protocol) error {
		p.pullConsumeOpts = pullConsumeOpts
		return nil
	}
}
