package pubsub

// Option is the function signature required to be considered an pubsub.Option.
type Option func(*Transport) error

// WithEncoding sets the encoding for pubsub transport.
func WithEncoding(encoding Encoding) Option {
	return func(t *Transport) error {
		t.Encoding = encoding
		return nil
	}
}

// WithSubscription sets the subscription for pubsub transport.
func WithSubscriptionID(subscriptionID string) Option {
	return func(t *Transport) error {
		t.subscriptionID = subscriptionID
		return nil
	}
}
