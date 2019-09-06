package kafka

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go-dev/kafka"
)

// Option is the function signature required to be considered an kafka.Option.
type Option func(*Transport) error

// WithEncoding sets the encoding for kafka transport.
func WithEncoding(encoding Encoding) Option {
	return func(t *Transport) error {
		t.Encoding = encoding
		return nil
	}
}

// WithDefaultEncodingSelector sets the encoding selection strategy for
// default encoding selections based on Event.
func WithDefaultEncodingSelector(fn EncodingSelector) Option {
	return func(t *Transport) error {
		if t == nil {
			return fmt.Errorf("http default encoding selector option can not set nil transport")
		}
		if fn != nil {
			t.DefaultEncodingSelectionFn = fn
			return nil
		}
		return fmt.Errorf("kafka fn for DefaultEncodingSelector was nil")
	}
}

// WithBinaryEncoding sets the encoding selection strategy for
// default encoding selections based on Event, the encoded event will be the
// given version in Binary form.
func WithBinaryEncoding() Option {
	return func(t *Transport) error {
		if t == nil {
			return fmt.Errorf("kafka binary encoding option can not set nil transport")
		}

		t.DefaultEncodingSelectionFn = DefaultBinaryEncodingSelectionStrategy
		return nil
	}
}

// WithStructuredEncoding sets the encoding selection strategy for
// default encoding selections based on Event, the encoded event will be the
// given version in Structured form.
func WithStructuredEncoding() Option {
	return func(t *Transport) error {
		if t == nil {
			return fmt.Errorf("kafka structured encoding option can not set nil transport")
		}

		t.DefaultEncodingSelectionFn = DefaultStructuredEncodingSelectionStrategy
		return nil
	}
}

func WithKafkaConfig(config *kafka.ConfigMap) Option {
	return func(t *Transport) error {
		t.config = config
		return nil
	}
}

// WithTopicID sets the topic ID for kafka transport.
func WithTopic(topic string) Option {
	return func(t *Transport) error {
		t.topic = topic
		return nil
	}
}

func WithAdminClient(ac *kafka.AdminClient) Option {
	return func(t *Transport) error {
		t.adminClient = ac
		return nil
	}
}
func WithProducer(p *kafka.Producer) Option {
	return func(t *Transport) error {
		t.producer = p
		return nil
	}
}
func WithConsumer(c *kafka.Consumer) Option {
	return func(t *Transport) error {
		t.consumer = c
		return nil
	}
}
