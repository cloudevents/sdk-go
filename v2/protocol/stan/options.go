package stan

import (
	"errors"

	"github.com/nats-io/stan.go"

	"github.com/cloudevents/sdk-go/v2/binding"
)

var ErrInvalidQueueName = errors.New("invalid queue name for QueueSubscriber")

// StanOptions is a helper function to group a variadic stan.Option into []stan.Option
func StanOptions(opts ...stan.Option) []stan.Option {
	return opts
}

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

// WithTransformer adds a transformer, which Sender uses while encoding a binding.Message to an stan.Message
func WithTransformer(transformer binding.Transformer) SenderOption {
	return func(s *Sender) error {
		s.Transformers = append(s.Transformers, transformer)
		return nil
	}
}

type MessageOption func(*Message) error

func WithManualAcks() MessageOption {
	return func(r *Message) error {
		r.manualAcks = true
		return nil
	}
}

type ReceiverOption func(*Receiver) error

func WithMessageOptions(opts ...MessageOption) ReceiverOption {
	return func(r *Receiver) error {
		r.messageOpts = opts
		return nil
	}
}

type ConsumerOption func(*Consumer) error

// WithQueueSubscriber configures the transport to create a queue subscription instead of a standard subscription.
func WithQueueSubscriber(queue string) ConsumerOption {
	return func(p *Consumer) error {
		if queue == "" {
			return ErrInvalidQueueName
		}
		p.Subscriber = &QueueSubscriber{queue}
		return nil
	}
}

// WithSubscriptionOptions sets options to configure the STAN subscription.
func WithSubscriptionOptions(opts ...stan.SubscriptionOption) ConsumerOption {
	return func(p *Consumer) error {
		p.subscriptionOptions = opts
		return nil
	}
}

// WithUnsubscribeOnClose configures the Protocol to unsubscribe subscriptions on close, rather than just closing.
// This causes durable subscriptions to be forgotten by the STAN service and recreated durable subscriptions will
// act like they are newly created.
func WithUnsubscribeOnClose() ConsumerOption {
	return func(p *Consumer) error {
		p.UnsubscribeOnClose = true
		return nil
	}
}

type SenderOption func(*Sender) error
