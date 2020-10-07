package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
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

type roundTripperTest struct {
	statusCodes  []int
	requestCount int
}

func (r *roundTripperTest) RoundTrip(req *http.Request) (*http.Response, error) {
	code := r.statusCodes[r.requestCount]
	r.requestCount++
	if code == -1 {
		time.Sleep(2 * time.Second)
		return nil, errors.New("timeout")
	}

	return &http.Response{StatusCode: code}, nil
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
