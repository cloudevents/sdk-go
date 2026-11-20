/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package kafka_franz

import (
	"context"
	"errors"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Option is the function signature required to be considered a kafka_franz.Option.
type Option func(*Protocol) error

// WithClient sets a kgo.Client instance to initialize the protocol directly.
func WithClient(client *kgo.Client) Option {
	return func(p *Protocol) error {
		if client == nil {
			return errors.New("the kgo.Client option must not be nil")
		}
		p.client = newKgoClient(client)
		return nil
	}
}

// WithClientOptions sets the options used to create a kgo.Client.
// When the protocol creates the client, it also enables manual offset control
// with kgo.DisableAutoCommit and kgo.BlockRebalanceOnPoll so ACKs drive commits.
func WithClientOptions(opts ...kgo.Opt) Option {
	return func(p *Protocol) error {
		if len(opts) == 0 {
			return errors.New("the kgo client options must not be empty")
		}
		p.clientOptions = append(p.clientOptions, opts...)
		return nil
	}
}

// WithSenderTopic sets the default topic used for produces when the context does not override it.
func WithSenderTopic(topic string) Option {
	return func(p *Protocol) error {
		if topic == "" {
			return errors.New("the producer topic option must not be empty")
		}
		p.producerDefaultTopic = topic
		return nil
	}
}

// Opaque key type used to store the Kafka message key.
type messageKeyType struct{}

var keyForMessageKey = messageKeyType{}

// WithMessageKey returns a new context with the given Kafka message key.
func WithMessageKey(ctx context.Context, messageKey []byte) context.Context {
	keyCopy := append([]byte(nil), messageKey...)
	return context.WithValue(ctx, keyForMessageKey, keyCopy)
}

// MessageKeyFrom looks in the given context and returns the message key if found, otherwise nil.
func MessageKeyFrom(ctx context.Context) []byte {
	c := ctx.Value(keyForMessageKey)
	if c == nil {
		return nil
	}
	s, ok := c.([]byte)
	if !ok {
		return nil
	}
	return append([]byte(nil), s...)
}
