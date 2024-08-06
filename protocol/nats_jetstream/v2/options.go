/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package nats_jetstream

import (
	"errors"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var ErrInvalidQueueName = errors.New("invalid queue name for QueueSubscriber")
var ErrNoConsumerConfig = errors.New("no consumer config was given")
var ErrNoJetstream = errors.New("no jetstream implementation provided")
var ErrMoreThanOneStream = errors.New("more than one stream for given filter subjects")
var ErrMoreThanOneConsumerConfig = errors.New("more than one consumer config given")

// ConsumerType - consumer types that have configurations defined in jetstream package
type ConsumerType int

const (
	ConsumerType_Unknown ConsumerType = iota
	ConsumerType_Ordinary
	ConsumerType_Ordered
)

// NatsOptions is a helper function to group a variadic nats.ProtocolOption into
// []nats.Option that can be used by either Sender, Consumer or Protocol
func NatsOptions(opts ...nats.Option) []nats.Option {
	return opts
}

// ProtocolOption is the function signature required to be considered an nats.ProtocolOption.
type ProtocolOption func(*Protocol) error

func WithConsumerOptions(opts ...ConsumerOption) ProtocolOption {
	return func(p *Protocol) error {
		p.consumerOptions = opts
		return nil
	}
}

func WithSenderOptions(opts ...SenderOption) ProtocolOption {
	return func(p *Protocol) error {
		p.senderOptions = opts
		return nil
	}
}

type SenderOption func(*Sender) error

// WithPublishOptions configures the Sender
func WithPublishOptions(publishOpts []jetstream.PublishOpt) SenderOption {
	return func(s *Sender) error {
		s.PublishOpts = publishOpts
		return nil
	}
}

type ConsumerOption func(*Consumer) error

// WithQueueSubscriber configures the Consumer to join a queue group when subscribing
func WithQueueSubscriber(queue string) ConsumerOption {
	return func(c *Consumer) error {
		if queue == "" {
			return ErrInvalidQueueName
		}
		c.Subscriber = &QueueSubscriber{Queue: queue}
		return nil
	}
}

// WithConsumerConfig configures the Consumer with the given config
func WithConsumerConfig(consumerConfig *jetstream.ConsumerConfig) ConsumerOption {
	return func(c *Consumer) error {
		if c.OrderedConsumerConfig != nil {
			return ErrMoreThanOneConsumerConfig
		}
		c.ConsumerConfig = consumerConfig
		return nil
	}
}

// WithOrderedConsumerConfig configures the Consumer with the given config
func WithOrderedConsumerConfig(orderedConsumerConfig *jetstream.OrderedConsumerConfig) ConsumerOption {
	return func(c *Consumer) error {
		if c.ConsumerConfig != nil {
			return ErrMoreThanOneConsumerConfig
		}
		c.OrderedConsumerConfig = orderedConsumerConfig
		return nil
	}
}

// WithPullConsumeOptions configures the Consumer with the given pullConsumeOpts
func WithPullConsumeOptions(pullConsumeOpt []jetstream.PullConsumeOpt) ConsumerOption {
	return func(c *Consumer) error {
		c.PullConsumeOpt = pullConsumeOpt
		return nil
	}
}
