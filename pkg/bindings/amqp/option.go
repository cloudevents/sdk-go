package amqp

import "github.com/cloudevents/sdk-go/pkg/binding"

// amqp.Sender options
type SenderOptionFunc func(sender *sender)

// Add a transformer, which Sender uses while encoding a binding.Message to an amqp.Message
func WithTransformer(transformer binding.TransformerFactory) SenderOptionFunc {
	return func(sender *sender) {
		sender.transformers = append(sender.transformers, transformer)
	}
}
