package test

import (
	"context"
	"io"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type ChanResponderResponse struct {
	Message binding.Message
	Result  protocol.Result
}

// ChanResponder implements Responder by receiving Messages from a channel and outputting the result in an output channel.
type ChanResponder struct {
	In  <-chan binding.Message
	Out chan<- ChanResponderResponse
}

func (r *ChanResponder) Respond(ctx context.Context) (binding.Message, protocol.ResponseFn, error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case m, ok := <-r.In:
		if !ok {
			return nil, nil, io.EOF
		}
		return m, func(ctx context.Context, message binding.Message, result protocol.Result) error {
			r.Out <- ChanResponderResponse{
				Message: message,
				Result:  result,
			}
			return nil
		}, nil
	}
}

func (r *ChanResponder) Receive(ctx context.Context) (binding.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case m, ok := <-r.In:
		if !ok {
			return nil, io.EOF
		}
		return m, nil
	}
}

func (r *ChanResponder) Close(ctx context.Context) error { return nil }

var _ protocol.Responder = (*ChanResponder)(nil)
