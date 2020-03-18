package amqp

import "pack.ag/amqp"

// Option is the function signature required to be considered an amqp.Option.
type Option func(*Transport) error

// WithEncoding sets the encoding for amqp transport.
func WithEncoding(encoding Encoding) Option {
	return func(t *Transport) error {
		t.Encoding = encoding
		return nil
	}
}

// WithConnOpt sets a connection option for amqp
func WithConnOpt(opt amqp.ConnOption) Option {
	return func(t *Transport) error {
		t.connOpts = append(t.connOpts, opt)
		return nil
	}
}

// WithConnSASLPlain sets SASLPlain connection option for amqp
func WithConnSASLPlain(username, password string) Option {
	return WithConnOpt(amqp.ConnSASLPlain(username, password))
}

// WithSessionOpt sets a session option for amqp
func WithSessionOpt(opt amqp.SessionOption) Option {
	return func(t *Transport) error {
		t.sessionOpts = append(t.sessionOpts, opt)
		return nil
	}
}

// WithSenderLinkOption sets a link option for amqp
func WithSenderLinkOption(opt amqp.LinkOption) Option {
	return func(t *Transport) error {
		t.senderLinkOpts = append(t.senderLinkOpts, opt)
		return nil
	}
}

// WithReceiverLinkOption sets a link option for amqp
func WithReceiverLinkOption(opt amqp.LinkOption) Option {
	return func(t *Transport) error {
		t.receiverLinkOpts = append(t.receiverLinkOpts, opt)
		return nil
	}
}
