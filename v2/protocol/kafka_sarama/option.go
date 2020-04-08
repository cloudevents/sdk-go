package kafka_sarama

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/binding"
)

// SenderOptionFunc is the type of kafka_sarama.Sender options
type SenderOptionFunc func(sender *Sender)

// WithTransformer adds a transformer, which Sender uses while encoding a binding.Message to a sarama.ProducerMessage
func WithTransformer(transformer binding.Transformer) SenderOptionFunc {
	return func(sender *Sender) {
		sender.transformers = append(sender.transformers, transformer)
	}
}

// ProtocolOptionFunc is the type of kafka_sarama.Protocol options
type ProtocolOptionFunc func(protocol *Protocol)

func WithReceiverGroupId(groupId string) ProtocolOptionFunc {
	return func(protocol *Protocol) {
		protocol.receiverGroupId = groupId
	}
}

func WithSenderContextDecorators(decorator func(context.Context) context.Context) ProtocolOptionFunc {
	return func(protocol *Protocol) {
		protocol.SenderContextDecorators = append(protocol.SenderContextDecorators, decorator)
	}
}
