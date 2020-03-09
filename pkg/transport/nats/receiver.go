package nats

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/transport"
	"github.com/nats-io/nats.go"
	"time"
)

var _ transport.Receiver

// receiver for nats, implements transport.Receiver
type receiver struct {
	conn    *nats.Conn
	subject string
	sub     *nats.Subscription
}

func (r *receiver) Receive(ctx context.Context) (binding.Message, error) {
	var err error
	if r.sub == nil {
		r.sub, err = r.conn.SubscribeSync(r.subject)
		if err != nil {
			return nil, err
		}
	}

	// TODO: allow to pass in a timeout
	timeout := time.Second * 60

	msg, err := r.sub.NextMsg(timeout)
	if err != nil {
		return nil, err
	}
	return NewMessage(msg), nil
}

func (r *receiver) Close(ctx context.Context) error {
	defer r.conn.Close()
	return r.sub.Unsubscribe()
}

// Create a new Receiver which wraps an amqp.Receiver in a binding.Receiver
func NewReceiver(conn *nats.Conn, subject string) transport.Receiver {
	return &receiver{conn: conn, subject: subject}
}
