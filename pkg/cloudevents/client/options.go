package client

import (
	"context"
	"fmt"
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

// WithOverrides adds a file based extension overrides defaulter to the end of
// the defaulter chain. the Overrides defaulter starts an active file watcher.
// To cancel the watcher, terminate the passed in context.
func WithOverrides(ctx context.Context, root string) Option {
	return func(c *ceClient) error {
		fn, err := NewDefaultOverrides(ctx, root)
		if err != nil {
			return err
		}
		c.eventDefaulterFns = append(c.eventDefaulterFns, fn)
		return nil
	}
}

// WithConverterFn defines the function the transport will use to delegate
// conversion of non-decodable messages.
func WithConverterFn(fn ConvertFn) Option {
	return func(c *ceClient) error {
		if fn == nil {
			return fmt.Errorf("client option was given an nil message converter")
		}
		if c.transport.HasConverter() {
			return fmt.Errorf("transport converter already set")
		}
		c.convertFn = fn
		c.transport.SetConverter(c)
		return nil
	}
}
