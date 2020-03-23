package gochan

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

const (
	defaultChanDepth = 20
)

// SendReceiver is a reference implementation for using the CloudEvents binding
// integration.
type SendReceiver struct {
	sender   protocol.Sender
	receiver protocol.Receiver
}

func New() *SendReceiver {
	ch := make(chan binding.Message, defaultChanDepth)

	return &SendReceiver{
		sender:   Sender(ch),
		receiver: Receiver(ch),
	}
}

func (s *SendReceiver) Send(ctx context.Context, in binding.Message) (err error) {
	return s.sender.Send(ctx, in)
}

func (r *SendReceiver) Receive(ctx context.Context) (binding.Message, error) {
	return r.receiver.Receive(ctx)
}
