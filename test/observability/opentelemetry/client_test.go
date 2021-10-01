/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package opentelemetry

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	otelObs "github.com/cloudevents/sdk-go/observability/opentelemetry/v2/client"
	obshttp "github.com/cloudevents/sdk-go/observability/opentelemetry/v2/http"
	"github.com/cloudevents/sdk-go/v2/client"
	event "github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/cloudevents/sdk-go/v2/types"
)

func TestTracedClientReceive(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		event          event.Event
		expectedResult protocol.Result
		ack            bool
	}{
		"simple binary v0.3": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV03{
						Type:   "unit.test.client",
						Source: *types.ParseURIRef("/unit/test/client"),
						Time:   &types.Timestamp{Time: now},
						ID:     "AABBCCDDEE",
					}.AsV03(),
				}
				_ = e.SetData(event.ApplicationJSON, &map[string]string{
					"sq":  "42",
					"msg": "hello",
				})
				return e
			}(),
			ack: true,
		},
		"receive with error": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV03{
						Type:   "unit.test.client",
						Source: *types.ParseURIRef("/unit/test/client"),
						Time:   &types.Timestamp{Time: now},
						ID:     "AABBCCDDEE",
					}.AsV03(),
				}
				_ = e.SetData(event.ApplicationJSON, &map[string]string{
					"sq":  "42",
					"msg": "hello",
				})
				return e
			}(),
			expectedResult: cehttp.NewResult(500, "some error happened within the receiver"),
			ack:            false,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			sr, tracer := configureOtelTestSdk()

			// creates and starts the receiver
			p, err := obshttp.NewObservedHTTP(cehttp.WithPort(0))
			require.NoError(t, err)
			c, err := client.New(p, client.WithObservabilityService(otelObs.NewOTelObservabilityService()))
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.TODO())
			go func() {
				require.NoError(t, c.StartReceiver(ctx, func(ctx context.Context, e event.Event) protocol.Result {
					return tc.expectedResult
				}))
			}()
			time.Sleep(5 * time.Millisecond) // let the server start

			target := fmt.Sprintf("http://localhost:%d", p.GetListeningPort())
			sender, err := otelObs.NewClientHTTP([]cehttp.Option{cehttp.WithTarget(target)}, []client.Option{})
			require.NoError(t, err)

			// act
			ctx, span := tracer.Start(ctx, "test-span")
			result := sender.Send(ctx, tc.event)
			span.End()

			require.Equal(t, tc.ack, protocol.IsACK(result))

			spans := sr.Ended()

			// 1 span from the test
			// 2 spans from sending the event (http client auto-instrumentation + obs service)
			// 2 spans from receiving the event (http client middleware + obs service)
			assert.Equal(t, 5, len(spans))

			if !tc.ack {
				// The span created by the observability service should have the error that came from the receiver fn
				obsSpan := spans[0]
				assert.Equal(t, codes.Error, obsSpan.Status().Code)
				assert.Equal(t, "500: some error happened within the receiver", obsSpan.Status().Description)
			}

			// Now stop the client
			cancel()
		})
	}
}

func TestTracingClientSend(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		event event.Event
		resp  *http.Response
		ack   bool
	}{
		"send with ok response": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV1{
						Type:   "unit.test.client",
						Source: *types.ParseURIRef("/unit/test/client"),
						Time:   &types.Timestamp{Time: now},
						ID:     "AABBCCDDEE",
					}.AsV1(),
				}
				_ = e.SetData(event.ApplicationJSON, &map[string]interface{}{
					"sq":  42,
					"msg": "hello",
				})
				return e
			}(),
			resp: &http.Response{
				StatusCode: http.StatusAccepted,
			},
			ack: true,
		},
		"send with error response": {
			event: func() event.Event {
				e := event.Event{
					Context: event.EventContextV1{
						Type:   "unit.test.client",
						Source: *types.ParseURIRef("/unit/test/client"),
						Time:   &types.Timestamp{Time: now},
						ID:     "AABBCCDDEE",
					}.AsV1(),
				}
				_ = e.SetData(event.ApplicationJSON, &map[string]interface{}{
					"sq":  42,
					"msg": "hello",
				})
				return e
			}(),
			resp: &http.Response{
				StatusCode: http.StatusBadRequest,
			},
			ack: false,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			sr, tracer := configureOtelTestSdk()
			handler := &fakeHandler{
				t:        t,
				response: tc.resp,
				requests: make([]*http.Request, 0),
			}
			server := httptest.NewServer(handler)
			defer server.Close()

			sender, err := otelObs.NewClientHTTP([]cehttp.Option{cehttp.WithTarget(server.URL)}, []client.Option{})
			require.NoError(t, err)

			// act
			ctx, span := tracer.Start(context.Background(), "test-span")
			result := sender.Send(ctx, tc.event)
			span.End()

			require.Equal(t, tc.ack, protocol.IsACK(result))

			spans := sr.Ended()

			// 1 span from the test
			// 2 spans from sending the event (http client auto-instrumentation + obs service)
			// 2 spans from receiving the event (http client middleware + obs service)
			assert.Equal(t, 3, len(spans))

			// get the traceparent header from the outgoing request
			r := handler.popRequest(t)
			if tp := r.Header.Get("traceparent"); tp == "" {
				t.Fatal("missing traceparent header")
			}

			// The request should have been sent with the last spanID (from the auto-instrumentation lib)
			ctx = prop.Extract(ctx, propagation.HeaderCarrier(r.Header))
			spanCtx := trace.SpanContextFromContext(ctx)
			assert.Equal(t, spans[0].SpanContext().TraceID(), spanCtx.TraceID())
			assert.Equal(t, spans[0].SpanContext().SpanID(), spanCtx.SpanID())
		})
	}
}

type fakeHandler struct {
	t        *testing.T
	response *http.Response
	requests []*http.Request
}

func (f *fakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Make a copy of the request.
	f.requests = append(f.requests, r)

	// Write the response.
	if f.response != nil {
		for h, vs := range f.response.Header {
			for _, v := range vs {
				w.Header().Add(h, v)
			}
		}
		w.WriteHeader(f.response.StatusCode)
		var buf bytes.Buffer
		if f.response.ContentLength > 0 {
			_, _ = buf.ReadFrom(f.response.Body)
			_, _ = w.Write(buf.Bytes())
		}
	} else {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(""))
	}
}

func (f *fakeHandler) popRequest(t *testing.T) *http.Request {
	if len(f.requests) == 0 {
		t.Error("Unable to pop request")
	}
	r := f.requests[0]
	f.requests = f.requests[1:]
	return r
}
