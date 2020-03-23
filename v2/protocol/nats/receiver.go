package nats

import (
	"context"
	"errors"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/nats-io/nats.go"
	"io"
	"sync"
)

var ErrSubscriptionAlreadyOpen = errors.New("subscription already open")

type msgErr struct {
	msg binding.Message
}

type Receiver struct {
	incoming chan msgErr
}

func NewReceiver() *Receiver {
	return &Receiver{
		incoming: make(chan msgErr),
	}
}

// MsgHandler implements nats.MsgHandler and publishes messages onto our internal incoming channel to be delivered
// via r.Receive(ctx)
func (r *Receiver) MsgHandler(msg *nats.Msg) {
	r.incoming <- msgErr{msg: NewMessage(msg)}
}

func (r *Receiver) Receive(ctx context.Context) (binding.Message, error) {
	select {
	case msgErr, ok := <-r.incoming:
		if !ok {
			return nil, io.EOF
		}
		return msgErr.msg, nil
	case <-ctx.Done():
		return nil, io.EOF
	}
}

type Consumer struct {
	Receiver

	Conn       *nats.Conn
	Subject    string
	Subscriber Subscriber

	sub       *nats.Subscription
	subMtx    sync.Mutex
	connOwned bool
}

func NewConsumer(url, subject string, natsOpts []nats.Option, opts ...ConsumerOption) (*Consumer, error) {
	conn, err := nats.Connect(url, natsOpts...)
	if err != nil {
		return nil, err
	}

	c, err := NewConsumerFromConn(conn, subject, opts...)
	if err != nil {
		conn.Close()
		return nil, err
	}

	c.connOwned = true

	return c, err
}

func NewConsumerFromConn(conn *nats.Conn, subject string, opts ...ConsumerOption) (*Consumer, error) {
	c := &Consumer{
		Receiver:   *NewReceiver(),
		Conn:       conn,
		Subject:    subject,
		Subscriber: &RegularSubscriber{},
	}

	err := c.applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Consumer) OpenInbound(ctx context.Context) error {
	var err error
	c.subMtx.Lock()
	if c.sub != nil && c.sub.IsValid() {
		return ErrSubscriptionAlreadyOpen
	}

	c.sub, err = c.Subscriber.Subscribe(c.Conn, c.Subject, c.MsgHandler)
	if err != nil {
		c.sub = nil
		c.subMtx.Unlock()
		return err
	}
	c.subMtx.Unlock()

	<-ctx.Done()

	return nil
}

// Receive consumes a message from the subscription
func (c *Consumer) Receive(ctx context.Context) (binding.Message, error) {
	return c.Receiver.Receive(ctx)
}

func (c *Consumer) Close(ctx context.Context) error {
	c.subMtx.Lock()
	defer c.subMtx.Unlock()

	if c.connOwned {
		defer c.Conn.Close()
	}

	if c.sub != nil {
		if err := c.sub.Unsubscribe(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Consumer) applyOptions(opts ...ConsumerOption) error {
	for _, fn := range opts {
		if err := fn(c); err != nil {
			return err
		}
	}
	return nil
}
