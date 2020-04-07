package context

import (
	"context"
	"testing"
	"time"
)

func TestRetryParams_Backoff(t *testing.T) {
	tests := map[string]struct {
		rp      *RetryParams
		ctx     context.Context
		tries   int
		wantErr bool
	}{
		"none 1": {
			ctx:     context.Background(),
			rp:      &RetryParams{Strategy: BackoffStrategyNone},
			tries:   1,
			wantErr: true,
		},
		"const 1": {
			ctx:   context.Background(),
			rp:    &RetryParams{Strategy: BackoffStrategyConstant, MaxTries: 10, Period: 1 * time.Nanosecond},
			tries: 5,
		},
		"linear 1": {
			ctx:   context.Background(),
			rp:    &RetryParams{Strategy: BackoffStrategyLinear, MaxTries: 10, Period: 1 * time.Nanosecond},
			tries: 1,
		},
		"exponential 1": {
			ctx:   context.Background(),
			rp:    &RetryParams{Strategy: BackoffStrategyExponential, MaxTries: 10, Period: 1 * time.Nanosecond},
			tries: 1,
		},
		"const timeout": {
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
				return ctx
			}(),
			rp:      &RetryParams{Strategy: BackoffStrategyConstant, MaxTries: 10, Period: 1 * time.Second},
			tries:   5,
			wantErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if err := tc.rp.Backoff(tc.ctx, tc.tries); (err != nil) != tc.wantErr {
				t.Errorf("Backoff() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestRetryParams_BackoffFor(t *testing.T) {
	tests := map[string]struct {
		rp    *RetryParams
		tries int
		want  time.Duration
	}{
		"none 1": {
			rp:    &RetryParams{Strategy: BackoffStrategyNone},
			tries: 1,
			want:  time.Duration(0),
		},
		"const 1": {
			rp:    &RetryParams{Strategy: BackoffStrategyConstant, MaxTries: 10, Period: 1 * time.Second},
			tries: 1,
			want:  1 * time.Second,
		},
		"linear 1": {
			rp:    &RetryParams{Strategy: BackoffStrategyLinear, MaxTries: 10, Period: 1 * time.Second},
			tries: 1,
			want:  1 * time.Second,
		},
		"exponential 1": {
			rp:    &RetryParams{Strategy: BackoffStrategyExponential, MaxTries: 10, Period: 1 * time.Second},
			tries: 1,
			want:  2 * time.Second, // 1 == 2^1
		},
		"none 5": {
			rp:    &RetryParams{Strategy: BackoffStrategyNone},
			tries: 5,
			want:  time.Duration(0),
		},
		"const 5": {
			rp:    &RetryParams{Strategy: BackoffStrategyConstant, MaxTries: 10, Period: 1 * time.Second},
			tries: 5,
			want:  1 * time.Second,
		},
		"linear 5": {
			rp:    &RetryParams{Strategy: BackoffStrategyLinear, MaxTries: 10, Period: 1 * time.Second},
			tries: 5,
			want:  5 * time.Second,
		},
		"exponential 5": {
			rp:    &RetryParams{Strategy: BackoffStrategyExponential, MaxTries: 10, Period: 1 * time.Second},
			tries: 5,
			want:  32 * time.Second, // 32 == 2^5
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tc.rp.BackoffFor(tc.tries); got != tc.want {
				t.Errorf("BackoffFor() = %v, want %v", got, tc.want)
			}
		})
	}
}
