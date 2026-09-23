/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	thttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"go.uber.org/zap"
	"net/http"
)

func NewHTTPReceiveHandler(ctx context.Context, p *thttp.Protocol, fn interface{}) (*EventReceiver, error) {
	invoker, err := newReceiveInvoker(fn, noopObservabilityService{}, nil, nil, false) //TODO(slinkydeveloper) maybe not nil?
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
	// The goroutine below must stay alive until r.p.ServeHTTP has returned, rather than
	// until the request context is cancelled.
	//
	// r.p.ServeHTTP hands the message over on the unbuffered Protocol.incoming channel and
	// blocks there until some consumer picks it up. Waiting on the request context lets this
	// goroutine exit -- on client disconnect -- while the message it was meant to pick up is
	// still pending. That permanently breaks the 1:1 pairing between ServeHTTP calls and
	// consumers: from then on every request is served by the *next* request's goroutine, so
	// each one blocks until further traffic arrives. See #1224.
	//
	// The cancellation cannot simply be dropped: r.p.ServeHTTP has paths that return without
	// sending anything (rate limiting, OPTIONS, GET) and Protocol.incoming is never closed, so
	// the goroutine would then block forever. Cancelling on return handles both cases -- while
	// a message is pending its producer has not returned yet, so at least one consumer is
	// guaranteed to still be waiting.
	waitCtx, stopWaiting := context.WithCancel(context.Background())
	defer stopWaiting()

	// Prepare to handle the message if there's one
	go func() {
		ctx := req.Context()
		msg, respFn, err := r.p.Respond(waitCtx)
		if err != nil {
			cecontext.LoggerFrom(context.TODO()).Debugw("failed to call Respond", zap.Error(err))
		} else if err := r.invoker.Invoke(ctx, msg, respFn); err != nil {
			cecontext.LoggerFrom(context.TODO()).Debugw("failed to call Invoke", zap.Error(err))
		}
	}()
	r.p.ServeHTTP(rw, req)
}
