package gochan

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

const (
	defaultChanDepth = 20
)

// Protocol is a reference implementation for using the CloudEvents binding
// integration.
type Protocol struct {
	sender   protocol.Sender
	receiver protocol.Receiver
}

func New() *Protocol {
	ch := make(chan binding.Message, defaultChanDepth)

	return &Protocol{
		sender:   ChanSender(ch),
		receiver: ChanReceiver(ch),
	}
}

func (s *Protocol) Send(ctx context.Context, in binding.Message) (err error) {
	return s.sender.Send(ctx, in)
}

func (r *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	return r.receiver.Receive(ctx)
}
