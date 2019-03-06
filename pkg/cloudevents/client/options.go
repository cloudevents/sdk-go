package client

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport/nats"
)

type Option func(*ceClient) error

// WithNATSEncoding sets the encoding for clients with NATS transport.
func WithNATSEncoding(encoding nats.Encoding) Option {
	return func(c *ceClient) error {
		if t, ok := c.transport.(*nats.Transport); ok {
			t.Encoding = encoding
			return nil
		}
		return fmt.Errorf("invalid NATS encoding client option received for transport type")
	}
}

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
