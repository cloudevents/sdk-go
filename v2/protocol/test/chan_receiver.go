package test

import (
	"context"
	"io"

	"github.com/cloudevents/sdk-go/v2/binding"
)

// ChanReceiver implements Receiver by receiving Messages from a channel.
type ChanReceiver <-chan binding.Message

func (r ChanReceiver) Receive(ctx context.Context) (binding.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case m, ok := <-r:
		if !ok {
			return nil, io.EOF
		}
		return m, nil
	}
}

func (r ChanReceiver) Close(ctx context.Context) error { return nil }
