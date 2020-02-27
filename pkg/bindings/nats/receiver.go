package nats

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/nats-io/nats.go"
	"io"
)

type Receiver struct {
	subscription *nats.Subscription
}

var _ binding.Receiver = (*Receiver)(nil)

func NewReceiver(subscription *nats.Subscription) *Receiver {
	return &Receiver{subscription: subscription}
}

func (r *Receiver) Receive(ctx context.Context) (binding.Message, error) {
	msg, err := r.subscription.NextMsgWithContext(ctx)
	if err == ctx.Err() {
		return nil, io.EOF
	} else if err != nil {
		return nil, err
	}

	return NewMessage(msg)
}
