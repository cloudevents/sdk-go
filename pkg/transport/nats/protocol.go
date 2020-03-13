package nats

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"time"

	"github.com/nats-io/nats.go"
)

// Protocol is a reference implementation for using the CloudEvents binding
// integration. Protocol acts as both a NATS client and a NATS handler.
type Protocol struct {
	Conn         *nats.Conn
	ConnOptions  []nats.Option
	NatsURL      string
	Subject      string
	Transformers binding.TransformerFactories

	subscription *nats.Subscription
}

// New creates a new NATS protocol.
func New(natsURL, subject string, opts ...Option) (*Protocol, error) {
	t := &Protocol{
		Subject:      subject,
		NatsURL:      natsURL,
		ConnOptions:  []nats.Option{},
		Transformers: make(binding.TransformerFactories, 0),
	}

	err := t.applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	err = t.connect()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Protocol) connect() error {
	var err error

	t.Conn, err = nats.Connect(t.NatsURL, t.ConnOptions...)

	return err
}

func (t *Protocol) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

// Send implements Sender.Send
func (t *Protocol) Send(ctx context.Context, in binding.Message) error {
	msg := &nats.Msg{}
	if err := WriteMsg(ctx, in, msg, t.Transformers); err != nil {
		return err
	}
	msg.Subject = t.Subject
	return t.Conn.PublishMsg(msg)
}

// Receive implements Receiver.Receive
func (t *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	var err error
	if t.subscription == nil {
		t.subscription, err = t.Conn.SubscribeSync(t.Subject)
		if err != nil {
			return nil, err
		}
	}

	// TODO: allow to pass in a timeout
	timeout := time.Second * 60

	msg, err := t.subscription.NextMsg(timeout)
	if err != nil {
		return nil, err
	}
	return NewMessage(msg), nil
}

// Close implements Closer.Close
func (t *Protocol) Close(ctx context.Context) error {
	defer t.Conn.Close()
	if t.subscription != nil {
		if err := t.subscription.Unsubscribe(); err != nil {
			return err
		}
		t.subscription = nil
	}
	return nil
}
