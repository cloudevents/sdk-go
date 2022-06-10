/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

func TestNew(t *testing.T) {
	dst := DefaultShutdownTimeout

	testCases := map[string]struct {
		opts    []Option
		want    *Protocol
		wantErr string
	}{
		"no options": {
			want: &Protocol{
				Client:          http.DefaultClient,
				ShutdownTimeout: dst,
				Port:            -1,
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			got, err := New(tc.opts...)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("Expected error '%s'. Actual '%v'", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreUnexported(Protocol{})); diff != "" {
				t.Errorf("unexpected diff (-want, +got) = %v", diff)
			}
		})
	}
}

func protocols(t *testing.T) []*Protocol {
	ps := make([]*Protocol, 1)

	p, err := New()
	if err != nil {
		t.Fatalf("Failed to create test Protocol: %s", err)
	}

	ps[0] = p

	return ps
}

func TestSend(t *testing.T) {
	testCases := map[string]struct {
		ctx     context.Context
		msg     binding.Message
		wantErr string
	}{
		"nil context": {
			wantErr: "nil Context",
		},
		"nil message": {
			ctx:     context.TODO(),
			wantErr: "nil Message",
		},
	}
	for n, tc := range testCases {
		for _, p := range protocols(t) {
			t.Run(n, func(t *testing.T) {
				err := p.Send(tc.ctx, tc.msg)
				if tc.wantErr != "" {
					if err == nil || err.Error() != tc.wantErr {
						t.Fatalf("Expected error '%s'. Actual '%v'", tc.wantErr, err)
					}
				} else if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			})
		}
	}
}

func TestRequest(t *testing.T) {
	testCases := map[string]struct {
		ctx     context.Context
		msg     binding.Message
		want    binding.Message
		wantErr string
	}{
		"nil context": {
			wantErr: "nil Context",
		},
		"nil message": {
			ctx:     context.TODO(),
			wantErr: "nil Message",
		},
	}
	for n, tc := range testCases {
		for _, p := range protocols(t) {
			t.Run(n, func(t *testing.T) {
				got, err := p.Request(tc.ctx, tc.msg)
				if tc.wantErr != "" {
					if err == nil || err.Error() != tc.wantErr {
						t.Fatalf("Expected error '%s'. Actual '%v'", tc.wantErr, err)
					}
				} else if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected diff (-want, +got) = %v", diff)
				}
			})
		}
	}
}

func TestReceive(t *testing.T) {
	testCases := map[string]struct {
		ctx     context.Context
		want    binding.Message
		wantErr string
	}{
		"nil context": {
			wantErr: "nil Context",
		},
		"timeout": {
			ctx:     newDoneContext(),
			wantErr: "EOF",
		},
	}
	for n, tc := range testCases {
		for _, p := range protocols(t) {
			t.Run(n, func(t *testing.T) {
				ReceiveTest(t, p, tc.ctx, nil, tc.want, tc.wantErr)
			})
		}
	}
}

func TestRespond(t *testing.T) {
	testCases := map[string]struct {
		ctx     context.Context
		want    binding.Message
		wantErr string
		resp    struct {
			ctx     context.Context
			msg     binding.Message
			result  protocol.Result
			wantErr string
		}
	}{
		"nil context": {
			wantErr: "nil Context",
		},
		"timeout context": {
			ctx:     newDoneContext(),
			wantErr: "EOF",
		},
		"non-expiring context": {
			ctx: context.Background(),
		},
	}
	for n, tc := range testCases {
		for _, p := range protocols(t) {
			t.Run(n, func(t *testing.T) {
				go func() {
					time.Sleep(time.Millisecond * 10)
					p.incoming <- msgErr{}
				}()

				got, fn, err := p.Respond(tc.ctx)
				if tc.wantErr != "" {
					assert.EqualError(t, err, tc.wantErr)
				} else {
					assert.NoError(t, err)
				}

				if got != nil {
					assert.NotNil(t, got, "Nil interface compares to nil")
				}

				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("unexpected diff (-want, +got) = %v", diff)
				}

				if fn != nil {
					err = fn(tc.resp.ctx, tc.resp.msg, tc.resp.result)
					if tc.resp.wantErr != "" {
						if err == nil || err.Error() != tc.resp.wantErr {
							t.Fatalf("Expected error '%s'. Actual '%v'", tc.resp.wantErr, err)
						}
					} else if err != nil {
						t.Fatalf("Unexpected error: %v", err)
					}
				}
			})
		}
	}
}

func TestServeHTTP_Receive(t *testing.T) {
	testCases := map[string]struct {
		// ServeHTTP
		rw  http.ResponseWriter
		req *http.Request
		// Receive
		want    binding.Message
		wantErr string
	}{
		"non-event": {
			rw:  httptest.NewRecorder(),
			req: httptest.NewRequest("POST", "http://unittest", nil),
		},
	}
	for n, tc := range testCases {
		for _, p := range protocols(t) {
			t.Run(n, func(t *testing.T) {
				go p.ServeHTTP(tc.rw, tc.req)
				rec := (tc.rw).(*httptest.ResponseRecorder)
				ReceiveTest(t, p, context.Background(), rec, tc.want, tc.wantErr)
			})
		}
	}
}

func ReceiveTest(t *testing.T, p *Protocol, ctx context.Context, rec *httptest.ResponseRecorder, want binding.Message, wantErr string) {
	got, err := p.Receive(ctx)
	if wantErr != "" {
		assert.EqualError(t, err, wantErr)

		if rec != nil {
			defer rec.Result().Body.Close()
			// TODO perform assertions on result if necessary
		}
	} else {
		assert.NoError(t, err)
	}

	if want == nil {
		require.Nil(t, want)
	} else {
		require.IsType(t, want, got)
	}
}

func TestServeHTTP_ReceiveWithLimiter(t *testing.T) {
	testCases := map[string]struct {
		limiter RateLimiter
		delay   time.Duration // client send

		// expect 405 if not rate limited (429) due to using GET to shorten test and
		// avoid internal response handling logic
		wantCodes []int
	}{
		"10rps limit, 5 requests, no delay, 405,405,405,405,405": {
			limiter:   newRateLimiterTest(10),
			delay:     0,
			wantCodes: []int{405, 405, 405, 405, 405},
		},
		"1rps limit, 5 requests, 100ms delay, 405,429,429,429,429": {
			limiter:   newRateLimiterTest(1),
			delay:     time.Millisecond * 100,
			wantCodes: []int{405, 429, 429, 429, 429},
		},
		"2rps limit, 4 requests, 0.5s delay, 405,200,200,200": {
			limiter:   newRateLimiterTest(2),
			delay:     time.Millisecond * 500,
			wantCodes: []int{405, 405, 405, 405},
		},
		// limiter disabled (backwards-compatibility)
		"no limit, 5 requests, no delay, 405,405,405,405,405": {
			limiter:   nil,
			delay:     0,
			wantCodes: []int{405, 405, 405, 405, 405},
		},
		// rate limit all
		"0rps limit, 5 requests, no delay, 429,429,429,429": {
			limiter:   newRateLimiterTest(0),
			delay:     time.Millisecond * 500,
			wantCodes: []int{429, 429, 429, 429},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			p, err := New(WithRateLimiter(tc.limiter))
			require.NoError(t, err, "create protocol")

			for _, code := range tc.wantCodes {
				time.Sleep(tc.delay)

				// using GET will give 405 if not rate limited and bypass any internal message
				// processing/async logic for faster and reproducible rate limit tests
				req := httptest.NewRequest("GET", "http://unittest", nil)
				rw := httptest.NewRecorder()

				p.ServeHTTP(rw, req)
				res := rw.Result()
				require.Equal(t, code, res.StatusCode)

				if res.StatusCode == 429 {
					retryAfterWant := tc.limiter.(*rateLimiterTest).resetSeconds
					retryAfterGot := res.Header.Get("Retry-After")
					require.Equal(t, strconv.Itoa(retryAfterWant), retryAfterGot)
				}
			}
		})
	}
}

type rateLimiterTest struct {
	limiter      *rate.Limiter
	resetSeconds int // defines Retry-After
}

func newRateLimiterTest(rps float64) RateLimiter {
	rl := rateLimiterTest{
		limiter:      rate.NewLimiter(rate.Limit(rps), int(rps)),
		resetSeconds: 2,
	}

	return &rl
}

func (rl *rateLimiterTest) Allow(_ context.Context, _ *http.Request) (bool, uint64, error) {
	if !rl.limiter.Allow() {
		return false, uint64(rl.resetSeconds), nil
	}
	return true, 0, nil
}

func (rl *rateLimiterTest) Close(_ context.Context) error {
	return nil
}

type roundTripperTest struct {
	sync.Mutex
	statusCodes  []int
	requestCount int
	delays       []time.Duration
}

func (r *roundTripperTest) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	defer func() {
		r.Lock()
		r.requestCount++
		r.Unlock()
	}()

	r.Lock()
	code := r.statusCodes[r.requestCount]
	delay := time.Duration(0)
	if r.delays != nil {
		delay = r.delays[r.requestCount]
	}
	r.Unlock()

	time.Sleep(delay)
	if code != 200 {
		http.Error(w, http.StatusText(code), code)
	}
}

func newDoneContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestDefaultIsRetriable(t *testing.T) {
	testCases := map[string]struct {
		statusCode  int
		isRetriable bool
	}{
		"400": {400, false},
		"404": {404, true},
		"408": {408, false},
		"413": {413, true},
		"425": {425, true},
		"429": {429, true},
		"500": {500, false},
		"502": {502, true},
		"503": {503, true},
		"504": {504, true},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			if got := defaultIsRetriableFunc(tc.statusCode); got != tc.isRetriable {
				t.Errorf("expected %v for %d but got %v", tc.isRetriable, tc.statusCode, got)
			}
		})
	}
}
