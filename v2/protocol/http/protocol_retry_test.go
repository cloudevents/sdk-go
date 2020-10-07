package http

import (
	"context"

	"github.com/stretchr/testify/require"

	"net/http"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

func TestRequestWithRetries_linear(t *testing.T) {
	dummyEvent := event.New()
	dummyMsg := binding.ToMessage(&dummyEvent)
	ctx := cecontext.WithTarget(context.Background(), "http://test")
	testCases := map[string]struct {
		// roundTripperTest
		statusCodes []int // -1 = timeout

		// Linear Backoff
		delay   time.Duration
		retries int

		// Wants
		wantResult       protocol.Result
		wantRequestCount int

		skipResults bool

		// Custom IsRetriable handler
		isRetriableFunc IsRetriable
	}{
		"no retries, ACK": {
			statusCodes: []int{200},
			retries:     0,
			wantResult: &RetriesResult{
				Result:  NewResult(200, "%w", protocol.ResultACK),
				Retries: 0,
			},
			wantRequestCount: 1,
		},
		"no retries, NACK": {
			statusCodes: []int{404},
			retries:     0,
			wantResult: &RetriesResult{
				Result:  NewResult(404, "%w", protocol.ResultNACK),
				Retries: 0,
			},
			wantRequestCount: 1,
		},
		"retries, no NACK": {
			statusCodes: []int{200},
			delay:       time.Nanosecond,
			retries:     3,
			wantResult: &RetriesResult{
				Result: NewResult(200, "%w", protocol.ResultACK),
			},
			wantRequestCount: 1,
		},
		"3 retries, 425, 200, ACK": {
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
		"no retries with default handler, 500, 200, ACK": {
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
		"3 retry with custom handler, 500, 500, 200, ACK": {
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
		"1 retry, 425, 429, 200, NACK": {
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
		"10 retries, 425, 429, 502, 503, 504, 200, ACK": {
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
		"retries, timeout, 200, ACK": {
			delay:       time.Nanosecond,
			statusCodes: []int{-1, 200},
			retries:     5,
			wantResult: &RetriesResult{
				Result:   NewResult(200, "%w", protocol.ResultACK),
				Retries:  1,
				Duration: time.Nanosecond,
				Attempts: nil, // skipping test as it contains internal http errors
			},
			wantRequestCount: 2,
			skipResults:      true,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			roundTripper := roundTripperTest{statusCodes: tc.statusCodes}
			opts := []Option{
				WithClient(http.Client{Timeout: time.Second}),
				WithRoundTripper(&roundTripper),
			}
			if tc.isRetriableFunc != nil {
				opts = append(opts, WithIsRetriableFunc(tc.isRetriableFunc))
			}

			p, err := New(opts...)
			if err != nil {
				t.Fatalf("no protocol")
			}
			ctxWithRetries := cecontext.WithRetriesLinearBackoff(ctx, tc.delay, tc.retries)
			_, got := p.Request(ctxWithRetries, dummyMsg)

			if roundTripper.requestCount != tc.wantRequestCount {
				t.Errorf("expected %d requests, got %d", tc.wantRequestCount, roundTripper.requestCount)
			}

			if tc.skipResults {
				got.(*RetriesResult).Attempts = nil
			}

			require.Equal(t, tc.wantResult.Error(), got.Error())
		})
	}
}
