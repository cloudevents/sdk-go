package http

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
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
				transformers:    binding.TransformerFactories{},
				Client:          http.DefaultClient,
				ShutdownTimeout: &dst,
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
	ps := make([]*Protocol, 0, 1)

	p, err := New()
	if err != nil {
		t.Fail()
	}
	ps = append(ps, p)
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
			ctx:     context.TODO(),
			wantErr: "EOF",
		},
	}
	for n, tc := range testCases {
		for _, p := range protocols(t) {
			t.Run(n, func(t *testing.T) {
				ReceiveTest(t, p, tc.ctx, tc.want, tc.wantErr)
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
		"timeout": {
			ctx:     context.TODO(),
			wantErr: "EOF",
		},
	}
	for n, tc := range testCases {
		for _, p := range protocols(t) {
			t.Run(n, func(t *testing.T) {
				if tc.ctx != nil {
					var done context.CancelFunc
					tc.ctx, done = context.WithDeadline(tc.ctx, time.Now().Add(time.Millisecond*10))
					defer done()
				}

				got, fn, err := p.Respond(tc.ctx)
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
		"nil": {
			wantErr: "unknown Message encoding",
		},
	}
	for n, tc := range testCases {
		for _, p := range protocols(t) {
			t.Run(n, func(t *testing.T) {
				go p.ServeHTTP(tc.rw, tc.req)
				ReceiveTest(t, p, context.Background(), tc.want, tc.wantErr)
			})
		}
	}
}

func ReceiveTest(t *testing.T, p *Protocol, ctx context.Context, want binding.Message, wantErr string) {
	if ctx != nil {
		var done context.CancelFunc
		ctx, done = context.WithTimeout(ctx, time.Millisecond*10)
		defer done()
	}

	got, err := p.Receive(ctx)
	if wantErr != "" {
		if err == nil || err.Error() != wantErr {
			t.Fatalf("Expected error '%s'. Actual '%v'", wantErr, err)
		}
	} else if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if want == nil {
		require.Nil(t, want)
	} else {
		require.IsType(t, want, got)
	}
}

func TestRequestWithRetries(t *testing.T) {
	dummyEvent := event.New()
	dummyMsg := binding.ToMessage(&dummyEvent)
	ctx := cecontext.WithTarget(context.Background(), "http://test")
	testCases := map[string]struct {
		// roundTripperTest
		statusCodes []int

		// Linear Backoff
		delay   time.Duration
		retries int

		// Wants
		wantResult       protocol.Result
		wantRequestCount int
	}{
		"425, 200, 3 retries": {
			statusCodes: []int{425, 200},
			delay:       time.Nanosecond,
			retries:     3,
			wantResult: &RetriesResult{
				Result:          NewResult(200, "%w", protocol.ResultACK),
				Retries:         1,
				RetriesDuration: time.Nanosecond,
				Results:         []protocol.Result{NewResult(425, "%w", protocol.ResultNACK)},
			},
			wantRequestCount: 2,
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			roundTripper := roundTripperTest{statusCodes: tc.statusCodes}

			p, err := New(WithRoundTripper(&roundTripper))
			if err != nil {
				t.Fail()
			}
			ctxWithRetries := cecontext.WithLinearBackoff(ctx, tc.delay, tc.retries)
			_, got := p.Request(ctxWithRetries, dummyMsg)

			if roundTripper.requestCount != tc.wantRequestCount {
				t.Errorf("expected %d requests, got %d", tc.wantRequestCount, roundTripper.requestCount)
			}

			if diff := cmp.Diff(tc.wantResult, got); diff != "" {
				t.Errorf("unexpected diff (-want, +got) = %v", diff)
			}
		})
	}
}

type roundTripperTest struct {
	statusCodes  []int
	requestCount int
}

func (r *roundTripperTest) RoundTrip(*http.Request) (*http.Response, error) {
	code := r.statusCodes[r.requestCount]
	r.requestCount++
	return &http.Response{StatusCode: code}, nil
}
