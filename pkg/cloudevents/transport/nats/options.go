package nats

type Option func(*Transport) error

// WithEncoding sets the encoding for NATS transport.
func WithEncoding(encoding Encoding) Option {
	return func(t *Transport) error {
		t.Encoding = encoding
		return nil
	}
}
