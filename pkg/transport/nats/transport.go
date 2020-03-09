package nats

import (
	"github.com/cloudevents/sdk-go/pkg/transport/bindings"

	"github.com/cloudevents/sdk-go/pkg/transport"
	"github.com/nats-io/nats.go"
)

// Transport adheres to transport.Transport.
var _ transport.Transport = (*Transport)(nil)

const (
	// TransportName is the name of this transport.
	TransportName = "NATS"
)

// Transport acts as both a NATS client and a NATS handler.
type Transport struct {
	bindings.BindingTransport

	Encoding    Encoding
	Conn        *nats.Conn
	ConnOptions []nats.Option
	NatsURL     string
	Subject     string
}

// New creates a new NATS transport.
func New(natsURL, subject string, opts ...Option) (*Transport, error) {
	t := &Transport{
		Subject:     subject,
		NatsURL:     natsURL,
		ConnOptions: []nats.Option{},
	}

	err := t.applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	err = t.connect()
	if err != nil {
		return nil, err
	}

	t.Sender = NewSender(t.Conn, t.Subject) // TODO: Allow sender options to be passed in.
	t.Receiver = NewReceiver(t.Conn, t.Subject)

	return t, nil
}

func (t *Transport) connect() error {
	var err error

	t.Conn, err = nats.Connect(t.NatsURL, t.ConnOptions...)

	return err
}

func (t *Transport) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}
