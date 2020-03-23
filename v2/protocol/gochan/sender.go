package gochan

import (
	"context"
	"errors"

	"github.com/cloudevents/sdk-go/v2/binding"
)

// ChanSender implements Sender by sending on a channel.
type ChanSender chan<- binding.Message

func (s ChanSender) Send(ctx context.Context, m binding.Message) (err error) {
	defer func() {
		err2 := m.Finish(err)
		if err == nil {
			err = err2
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s <- m:
		return nil
	}
}

func (s ChanSender) Close(ctx context.Context) (err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("send on closed channel")
		}
	}()
	close(s)
	return nil
}
