package nats

import "github.com/cloudevents/sdk-go/pkg/binding"

type SenderOptionFunc func(sender *Sender)

func WithTranscoder(factory binding.TransformerFactory) SenderOptionFunc {
	return func(sender *Sender) {
		sender.transformerFactories = append(sender.transformerFactories, factory)
	}
}
