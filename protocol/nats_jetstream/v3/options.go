/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// ProtocolOption provides a way to configure the protocol
type ProtocolOption func(*Protocol) error

// WithURL creates a connection to be used in the protocol sender and receiver.
// This option is mutually exclusive with WithConnection.
func WithURL(url string) ProtocolOption {
	return func(p *Protocol) error {
		p.url = url
		return nil
	}
}

// WithNatsOptions can be used together with WithURL() to specify NATS connection options
func WithNatsOptions(natsOpts []nats.Option) ProtocolOption {
	return func(p *Protocol) error {
		p.natsOpts = natsOpts
		return nil
	}
}

// WithConnection uses the provided connection in the protocol sender and receiver
// This option is mutually exclusive with WithURL.
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
		p.consumerConfig = consumerConfig
		return nil
	}
}

// WithOrderedConsumerConfig creates a ordered consumer used in the protocol receiver.
// This option is mutually exclusive with WithConsumerConfig.
func WithOrderedConsumerConfig(orderedConsumerConfig *jetstream.OrderedConsumerConfig) ProtocolOption {
	return func(p *Protocol) error {
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
