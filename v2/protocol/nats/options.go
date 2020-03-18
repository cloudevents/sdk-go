package nats

import (
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/nats-io/nats.go"
)

// Option is the function signature required to be considered an nats.Option.
type Option func(*Protocol) error

// WithConnOptions supplies NATS connection options that will be used when setting
// up the internal NATS connection
func WithConnOptions(opts ...nats.Option) Option {
	return func(t *Protocol) error {
		for _, o := range opts {
			t.ConnOptions = append(t.ConnOptions, o)
		}

		return nil
	}
}

// Add a transformer, which Protocol uses while encoding a binding.Message to an nats.Message
func WithTransformer(transformer binding.TransformerFactory) Option {
	return func(p *Protocol) error {
		p.Transformers = append(p.Transformers, transformer)
		return nil
	}
}
