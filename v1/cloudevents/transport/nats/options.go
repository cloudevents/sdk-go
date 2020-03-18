package nats

import (
	"github.com/nats-io/nats.go"
)

// Option is the function signature required to be considered an nats.Option.
type Option func(*Transport) error

// WithEncoding sets the encoding for NATS transport.
func WithEncoding(encoding Encoding) Option {
	return func(t *Transport) error {
		t.Encoding = encoding
		return nil
	}
}

// WithConnOptions supplies NATS connection options that will be used when setting
// up the internal NATS connection
func WithConnOptions(opts ...nats.Option) Option {
	return func(t *Transport) error {
		for _, o := range opts {
			t.ConnOptions = append(t.ConnOptions, o)
		}

		return nil
	}
}
