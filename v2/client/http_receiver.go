/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"net/http"

	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/protocol"
	thttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"go.uber.org/zap"
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
	// Deliver each request to the invoker inline (see Protocol.ServeHTTPWithHandler),
	// so a request cancelled mid-flight cannot orphan a concurrent request's
	// delivery and stall the receiver.
	r.p.ServeHTTPWithHandler(rw, req, func(ctx context.Context, m binding.Message, respFn protocol.ResponseFn) error {
		if err := r.invoker.Invoke(ctx, m, respFn); err != nil {
			cecontext.LoggerFrom(ctx).Debugw("failed to call Invoke", zap.Error(err))
			return err
		}
		return nil
	})
}
