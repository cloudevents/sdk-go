package http

import "github.com/cloudevents/sdk-go/pkg/binding"

type SenderOptionFunc func(sender *Sender)

func ForceBinary() SenderOptionFunc {
	return func(sender *Sender) {
		sender.forceBinary = true
	}
}

func ForceStructured() SenderOptionFunc {
	return func(sender *Sender) {
		sender.forceStructured = true
	}
}

func WithTranscoder(factory binding.TransformerFactory) SenderOptionFunc {
	return func(sender *Sender) {
		sender.transformerFactories = append(sender.transformerFactories, factory)
	}
}
