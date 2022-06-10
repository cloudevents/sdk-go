/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

func TestRequestWithRetries_linear(t *testing.T) {
	testCases := map[string]struct {
		event event.Event
		// roundTripperTest
		statusCodes []int // -1 = timeout

		// retry configuration
		delay   time.Duration
		retries int

		// http server
		respDelay []time.Duration // slice maps to []statusCodes, 0 for no delay

		// Wants
		wantResult       protocol.Result
		wantRequestCount int

		skipResults bool

		// Custom IsRetriable handler
		isRetriableFunc IsRetriable
	}{
		"no retries, no event body, ACK": {
			event:       newEvent(t, "", nil),
			statusCodes: []int{200},
			retries:     0,
			wantResult: &RetriesResult{
				Result:  NewResult(200, "%w", protocol.ResultACK),
				Retries: 0,
			},
			wantRequestCount: 1,
		},
		"no retries, no event body, NACK": {
			event:       newEvent(t, "", nil),
			statusCodes: []int{404},
			retries:     0,
			wantResult: &RetriesResult{
				Result:  NewResult(404, "%w", protocol.ResultNACK),
				Retries: 0,
			},
			wantRequestCount: 1,
		},
		"no retries, with default handler, with event body, 500, 200, ACK": {
			event:       newEvent(t, event.ApplicationJSON, "hello world"),
			statusCodes: []int{500, 200},
			delay:       time.Nanosecond,
			retries:     3,
			wantResult: &RetriesResult{
				Result:   NewResult(500, "%w", protocol.ResultNACK),
				Duration: time.Nanosecond,
				Attempts: []protocol.Result{NewResult(500, "%w", protocol.ResultNACK)},
			},
			wantRequestCount: 1,
		},
		"3 retries, no event body, 425, 200, ACK": {
			event:       newEvent(t, "", nil),
			statusCodes: []int{425, 200},
			delay:       time.Nanosecond,
			retries:     3,
			wantResult: &RetriesResult{
				Result:   NewResult(200, "%w", protocol.ResultACK),
				Retries:  1,
				Duration: time.Nanosecond,
				Attempts: []protocol.Result{NewResult(425, "%w", protocol.ResultNACK)},
			},
			wantRequestCount: 2,
		},
		"3 retries, with event body, 503, 503, 503, NACK": {
			event:       newEvent(t, event.ApplicationJSON, map[string]string{"hello": "world"}),
			delay:       time.Nanosecond,
			statusCodes: []int{503, 503, 503, 503},
			retries:     3,
			wantResult: &RetriesResult{
				Result:   NewResult(503, "%w", protocol.ResultNACK),
				Retries:  3,
				Duration: time.Nanosecond,
				Attempts: []protocol.Result{
					NewResult(503, "%w", protocol.ResultNACK),
					NewResult(503, "%w", protocol.ResultNACK),
					NewResult(503, "%w", protocol.ResultNACK),
				},
			},
			wantRequestCount: 4,
			skipResults:      true,
		},
		"3 retries, with custom handler, with event body, 500, 500, 200, ACK": {
			event:       newEvent(t, event.ApplicationJSON, map[string]string{"hello": "world"}),
			statusCodes: []int{500, 500, 200},
			delay:       time.Nanosecond,
			retries:     3,
			wantResult: &RetriesResult{
				Result:   NewResult(200, "%w", protocol.ResultACK),
				Duration: time.Nanosecond,
				Retries:  2,
				Attempts: []protocol.Result{
					NewResult(500, "%w", protocol.ResultNACK),
					NewResult(500, "%w", protocol.ResultNACK),
				},
			},
			wantRequestCount: 3,
			isRetriableFunc:  func(sc int) bool { return sc == 500 },
		},
		"1 retry, no event body, 425, 429, 200, NACK": {
			event:       newEvent(t, "", nil),
			statusCodes: []int{425, 429, 200},
			delay:       time.Nanosecond,
			retries:     1,
			wantResult: &RetriesResult{
				Result:   NewResult(429, "%w", protocol.ResultNACK),
				Retries:  1,
				Duration: time.Nanosecond,
				Attempts: []protocol.Result{NewResult(425, "%w", protocol.ResultNACK)},
			},
			wantRequestCount: 2,
		},
		"10 retries, with event body, 425, 429, 502, 503, 504, 200, ACK": {
			event:       newEvent(t, event.ApplicationJSON, map[string]string{"hello": "world"}),
			statusCodes: []int{425, 429, 502, 503, 504, 200},
			delay:       time.Nanosecond,
			retries:     10,
			wantResult: &RetriesResult{
				Result:  NewResult(200, "%w", protocol.ResultACK),
				Retries: 5,
				Attempts: []protocol.Result{
					NewResult(425, "%w", protocol.ResultNACK),
					NewResult(429, "%w", protocol.ResultNACK),
					NewResult(502, "%w", protocol.ResultNACK),
					NewResult(503, "%w", protocol.ResultNACK),
					NewResult(504, "%w", protocol.ResultNACK),
				},
			},
			wantRequestCount: 6,
		},
		"5 retries, with event body, timeout, 200, ACK": {
			event:       newEvent(t, event.ApplicationJSON, map[string]string{"hello": "world"}),
			delay:       time.Millisecond * 500,
			statusCodes: []int{200, 200}, // client will time out before first 200
			retries:     5,
			respDelay:   []time.Duration{time.Second, 0},
			wantResult: &RetriesResult{
				Result:   NewResult(200, "%w", protocol.ResultACK),
				Retries:  1,
				Duration: time.Millisecond * 500,
				Attempts: nil, // skipping test as it contains internal http errors
			},
			wantRequestCount: 2,
			skipResults:      true,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			mockSrv := &roundTripperTest{
				statusCodes: tc.statusCodes,
				delays:      tc.respDelay,
			}
			srv := httptest.NewServer(mockSrv)
			defer srv.Close()

			opts := []Option{
				WithClient(http.Client{Timeout: time.Second}),
			}
			if tc.isRetriableFunc != nil {
				opts = append(opts, WithIsRetriableFunc(tc.isRetriableFunc))
			}

			p, err := New(opts...)
			if err != nil {
				t.Fatalf("no protocol")
			}

			ctx := cecontext.WithTarget(context.Background(), srv.URL)
			ctxWithRetries := cecontext.WithRetriesLinearBackoff(ctx, tc.delay, tc.retries)

			dummyMsg := binding.ToMessage(&tc.event)
			_, got := p.Request(ctxWithRetries, dummyMsg)

			srvCount := func() int {
				mockSrv.Lock()
				defer mockSrv.Unlock()
				return mockSrv.requestCount
			}
			assert.Equal(t, tc.wantRequestCount, srvCount())

			if tc.skipResults {
				got.(*RetriesResult).Attempts = nil
			}

			require.Equal(t, tc.wantResult.Error(), got.Error())
		})
	}
}

func newEvent(t *testing.T, encoding string, body interface{}) event.Event {
	e := event.New()
	if body != nil {
		err := e.SetData(encoding, body)
		require.NoError(t, err)
	}

	return e
}
