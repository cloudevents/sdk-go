package kafka_sarama

import (
	"context"

	"github.com/cloudevents/sdk-go/pkg/binding"
)

// kafka_sarama.Sender options
type SenderOptionFunc func(sender *Sender)

// Add a transformer, which Sender uses while encoding a binding.Message to a sarama.ProducerMessage
func WithTransformer(transformer binding.TransformerFactory) SenderOptionFunc {
	return func(sender *Sender) {
		sender.transformers = append(sender.transformers, transformer)
	}
}

// kafka_sarama.Protocol options
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
