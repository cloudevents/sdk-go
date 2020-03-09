package kafka_sarama

import "github.com/cloudevents/sdk-go/pkg/binding"

// kafka_sarama.Sender options
type SenderOptionFunc func(sender *Sender)

// Add a transformer, which Sender uses while encoding a binding.Message to a sarama.ProducerMessage
func WithTransformer(transformer binding.TransformerFactory) SenderOptionFunc {
	return func(sender *Sender) {
		sender.transformers = append(sender.transformers, transformer)
	}
}
