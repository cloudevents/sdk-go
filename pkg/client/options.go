package client

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/binding"
)

// Option is the function signature required to be considered an client.Option.
type Option func(*ceClient) error

// WithEventDefaulter adds an event defaulter to the end of the defaulter chain.
func WithEventDefaulter(fn EventDefaulter) Option {
	return func(c *ceClient) error {
		if fn == nil {
			return fmt.Errorf("client option was given an nil event defaulter")
		}
		c.eventDefaulterFns = append(c.eventDefaulterFns, fn)
		return nil
	}
}

func WithForceBinary() Option {
	return func(c *ceClient) error {
		c.outboundContextDecorators = append(c.outboundContextDecorators, binding.WithForceBinary)
		return nil
	}
}

func WithForceStructured() Option {
	return func(c *ceClient) error {
		c.outboundContextDecorators = append(c.outboundContextDecorators, binding.WithForceStructured)
		return nil
	}
}

// WithUUIDs adds DefaultIDToUUIDIfNotSet event defaulter to the end of the
// defaulter chain.
func WithUUIDs() Option {
	return func(c *ceClient) error {
		c.eventDefaulterFns = append(c.eventDefaulterFns, DefaultIDToUUIDIfNotSet)
		return nil
	}
}

// WithTimeNow adds DefaultTimeToNowIfNotSet event defaulter to the end of the
// defaulter chain.
func WithTimeNow() Option {
	return func(c *ceClient) error {
		c.eventDefaulterFns = append(c.eventDefaulterFns, DefaultTimeToNowIfNotSet)
		return nil
	}
}

// WithoutTracePropagation disables automatic trace propagation via
// the distributed tracing extension.
func WithoutTracePropagation() Option {
	return func(c *ceClient) error {
		//c.disableTracePropagation = true
		return nil
	}
}
