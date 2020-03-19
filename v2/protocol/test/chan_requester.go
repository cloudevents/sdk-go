package test

import (
	"context"
	"errors"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type ChanRequester struct {
	Ch    chan<- binding.Message
	Reply func(message binding.Message) (binding.Message, error)
}

func (s *ChanRequester) Send(ctx context.Context, m binding.Message) (err error) {
	defer func() {
		err2 := m.Finish(err)
		if err == nil {
			err = err2
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.Ch <- m:
		return nil
	}
}

func (s *ChanRequester) Request(ctx context.Context, m binding.Message) (res binding.Message, err error) {
	defer func() {
		err2 := m.Finish(err)
		if err == nil {
			err = err2
		}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.Ch <- m:
		return s.Reply(m)
	}
}

func (s *ChanRequester) Close(ctx context.Context) (err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("trying to close a closed ChanSender")
		}
	}()
	close(s.Ch)
	return nil
}

var _ protocol.Requester = (*ChanRequester)(nil)
