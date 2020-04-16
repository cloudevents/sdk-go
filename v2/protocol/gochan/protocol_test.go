package gochan

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
	"time"
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
