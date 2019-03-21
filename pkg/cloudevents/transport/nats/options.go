package nats

// Option is the function signature required to be considered an nats.Option.
type Option func(*Transport) error

// WithEncoding sets the encoding for NATS transport.
func WithEncoding(encoding Encoding) Option {
	return func(t *Transport) error {
		t.Encoding = encoding
		return nil
	}
}
