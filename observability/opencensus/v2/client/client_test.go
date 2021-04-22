/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lightstep/tracecontext.go/traceparent"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/trace"

	obshttp "github.com/cloudevents/sdk-go/observability/opencensus/v2/http"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/cloudevents/sdk-go/v2/types"
)

func simpleTracingBinaryClient(t *testing.T, target string) client.Client {
	p, err := obshttp.NewObservedHTTP(cehttp.WithTarget(target))
	require.NoError(t, err)

	c, err := client.New(p, client.WithObservabilityService(New()))
	require.NoError(t, err)
	return c
}

func TestTracedClientReceive(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		event event.Event
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
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			spanContexts := make(chan trace.SpanContext, 1)

			p, err := obshttp.NewObservedHTTP(cehttp.WithPort(0))
			require.NoError(t, err)
			c, err := client.New(p, client.WithObservabilityService(New()))
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.TODO())
			go func() {
				require.NoError(t, c.StartReceiver(ctx, func(ctx context.Context, e event.Event) (*event.Event, protocol.Result) {
					span := trace.FromContext(ctx)
					spanContexts <- span.SpanContext()
					return nil, nil
				}))
			}()
			time.Sleep(5 * time.Millisecond) // let the server start

			target := fmt.Sprintf("http://localhost:%d", p.GetListeningPort())
			sender := simpleTracingBinaryClient(t, target)

			ctx, span := trace.StartSpan(context.TODO(), "test-span")
			result := sender.Send(ctx, tc.event)
			span.End()

			require.True(t, protocol.IsACK(result))

			got := <-spanContexts

			if span.SpanContext().TraceID != got.TraceID {
				t.Errorf("unexpected traceID. want: %s, got %s", span.SpanContext().TraceID, got.TraceID)
			}

			// Now stop the client
			cancel()
		})
	}
}

func TestTracingClientSend(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		c      func(t *testing.T, target string) client.Client
		event  event.Event
		resp   *http.Response
		sample bool
	}{
		"send unsampled": {
			c: simpleTracingBinaryClient,
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
		},
		"send sampled": {
			c: simpleTracingBinaryClient,
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
			sample: true,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			handler := &fakeHandler{
				t:        t,
				response: tc.resp,
				requests: make([]requestValidation, 0),
			}
			server := httptest.NewServer(handler)
			defer server.Close()

			c := tc.c(t, server.URL)

			var sampler trace.Sampler
			if tc.sample {
				sampler = trace.AlwaysSample()
			} else {
				sampler = trace.NeverSample()
			}
			ctx, span := trace.StartSpan(context.TODO(), "test-span", trace.WithSampler(sampler))
			sc := span.SpanContext()

			result := c.Send(ctx, tc.event)
			span.End()

			if !protocol.IsACK(result) {
				t.Fatalf("failed to send event: %s", result)
			}

			rv := handler.popRequest(t)

			var got traceparent.TraceParent
			if tp := rv.Headers.Get("Traceparent"); tp == "" {
				t.Fatal("missing traceparent header")
			} else {
				var err error
				got, err = traceparent.ParseString(tp)
				if err != nil {
					t.Fatalf("invalid traceparent: %s", err)
				}
			}
			if got.TraceID != sc.TraceID {
				t.Errorf("unexpected trace id: want %s got %s", sc.TraceID, got.TraceID)
			}
			if got.Flags.Recorded != tc.sample {
				t.Errorf("unexpected recorded flag: want %t got %t", tc.sample, got.Flags.Recorded)
			}
		})
	}
}

type requestValidation struct {
	Host    string
	Headers http.Header
	Body    []byte
}

type fakeHandler struct {
	t        *testing.T
	response *http.Response
	requests []requestValidation
}

func (f *fakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Make a copy of the request.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		f.t.Error("failed to read the request body")
	}
	f.requests = append(f.requests, requestValidation{
		Host:    r.Host,
		Headers: r.Header,
		Body:    body,
	})

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

func (f *fakeHandler) popRequest(t *testing.T) requestValidation {
	if len(f.requests) == 0 {
		t.Error("Unable to pop request")
	}
	rv := f.requests[0]
	f.requests = f.requests[1:]
	return rv
}
