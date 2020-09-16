package client

import (
	"context"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"go.uber.org/zap"
	"net/http"
	"sync"

	thttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

func NewHTTPReceiveHandler(ctx context.Context, p *thttp.Protocol, fn interface{}) (*EventReceiver, error) {
	invoker, err := newReceiveInvoker(fn)
	if err != nil {
		return nil, err
	}

	return &EventReceiver{
		p:       p,
		invoker: invoker,
	}, nil
}

type EventReceiver struct {
	p       *thttp.Protocol
	invoker Invoker
}

func (r *EventReceiver) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		r.p.ServeHTTP(rw, req)
		wg.Done()
	}()

	ctx := req.Context()
	msg, respFn, err := r.p.Respond(ctx)
	if err != nil {
		cecontext.LoggerFrom(context.TODO()).Debugw("failed to call Respond", zap.Error(err))
	} else if err := r.invoker.Invoke(ctx, msg, respFn); err != nil {
		cecontext.LoggerFrom(context.TODO()).Debugw("failed to call Invoke", zap.Error(err))
	}
	// Block until ServeHTTP has returned
	wg.Wait()
}
