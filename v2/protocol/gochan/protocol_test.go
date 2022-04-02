/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package gochan

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	got := New()
	assert.NotNil(t, got)
}

func protocols(t *testing.T) []*SendReceiver {
	return []*SendReceiver{New()}
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
			wantErr: io.EOF.Error(),
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

func TestSendReceive(t *testing.T) {
	testCases := map[string]struct {
		sendErr    string
		want       binding.Message
		receiveErr string
	}{
		"nil": {
			sendErr:    "nil Message",
			receiveErr: io.EOF.Error(),
		},
		"empty event": {
			want: func() binding.Message {
				e := event.New()
				return binding.ToMessage(&e)
			}(),
		},
		"min event": {
			want: func() binding.Message {
				e := event.New()
				e.SetSource("unittest/")
				e.SetType("unit.test")
				e.SetID("unit-test")
				return binding.ToMessage(&e)
			}(),
		},
	}
	for n, tc := range testCases {
		for _, p := range protocols(t) {
			t.Run(n, func(t *testing.T) {
				sendErrCh := make(chan error)
				go func() {
					sendErrCh <- p.Send(context.Background(), tc.want)
				}()

				ReceiveTest(t, p, context.Background(), tc.want, tc.receiveErr)

				err := <-sendErrCh
				wantErr := tc.sendErr
				if wantErr != "" {
					if err == nil || err.Error() != wantErr {
						t.Fatalf("Expected error '%s'. Actual '%v'", wantErr, err)
					}
				} else if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			})
		}
	}
}

func TestSendCloser(t *testing.T) {
	testCases := map[string]struct {
		numSend            int
		numReceivePreClose int
		numClose           int // defaults to 1
		wantErr            bool
	}{
		"closes none pending": {
			numSend:            1,
			numReceivePreClose: 1,
		},
		"closes still delivers pending": {
			numSend:            2,
			numReceivePreClose: 1,
		},
		"errors on double close": {
			numClose: 2,
			wantErr:  true,
		},
	}
	for n, tc := range testCases {
		for _, p := range protocols(t) {
			t.Run(n, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
				defer cancel()
				for i := 0; i < tc.numSend; i++ {
					e := event.New()
					if err := p.Send(ctx, binding.ToMessage(&e)); err != nil {
						t.Fatalf("failed to send to protocol: %v", err)
					}
				}

				for i := 0; i < tc.numReceivePreClose; i++ {
					_, err := p.Receive(ctx)
					if err != nil {
						t.Fatalf("failed to receive from protocol: %v", err)
					}
				}

				if tc.numClose == 0 {
					tc.numClose = 1
				}

				var err error
				for i := 0; i < tc.numClose; i++ {
					err = p.Close(ctx)
				}
				if tc.wantErr != (err != nil) {
					t.Fatalf("failed to close channel, wantErr = %v, got = %v", tc.wantErr, err)
				}

				for i := 0; i < tc.numSend-tc.numReceivePreClose; i++ {
					_, err := p.Receive(ctx)
					if err != nil {
						t.Fatalf("failed to receive from protocol: %v", err)
					}
				}

				if _, err = p.Receive(ctx); err != io.EOF {
					t.Fatalf("expected protocol to be closed but got err = %v", err)
				}
			})
		}
	}
}

func ReceiveTest(t *testing.T, p *SendReceiver, ctx context.Context, want binding.Message, wantErr string) {
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

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected diff (-want, +got) = %v", diff)
	}
}
