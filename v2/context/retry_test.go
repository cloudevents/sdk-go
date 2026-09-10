/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"context"
	"math"
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
		"exponential jitter 1": {
			ctx:   context.Background(),
			rp:    &RetryParams{Strategy: BackoffStrategyExponentialWithJitter, MaxTries: 10, Period: 1 * time.Nanosecond},
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

func TestRetryParams_BackoffFor_ExponentialWithJitter(t *testing.T) {
	for _, period := range []time.Duration{0, -1, -time.Second} {
		rp := &RetryParams{
			Strategy: BackoffStrategyExponentialWithJitter,
			MaxTries: 10,
			Period:   period,
		}
		if got := rp.BackoffFor(1); got != 0 {
			t.Errorf("BackoffFor() with period %v = %v, want 0", period, got)
		}
	}

	basePeriod := 10 * time.Millisecond
	rp := &RetryParams{
		Strategy: BackoffStrategyExponentialWithJitter,
		MaxTries: 10,
		Period:   basePeriod,
	}
	for tries := 1; tries <= 5; tries++ {
		ceiling := time.Duration(float64(basePeriod) * math.Exp2(float64(tries)))
		for i := 0; i < 50; i++ {
			got := rp.BackoffFor(tries)
			if got < 0 || got >= ceiling {
				t.Errorf("BackoffFor(%d) = %v, want in [0, %v)", tries, got, ceiling)
			}
		}
	}

	for _, tries := range []int{64, 1000} {
		for i := 0; i < 10; i++ {
			got := rp.BackoffFor(tries)
			if got < 0 {
				t.Errorf("BackoffFor(%d) = %v, want >= 0", tries, got)
			}
		}
	}

	ctx := context.Background()
	for _, period := range []time.Duration{0, -time.Millisecond} {
		rp := &RetryParams{
			Strategy: BackoffStrategyExponentialWithJitter,
			MaxTries: 10,
			Period:   period,
		}
		if err := rp.Backoff(ctx, 1); err != nil {
			t.Errorf("Backoff() with period %v error = %v, want nil", period, err)
		}
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	rpZero := &RetryParams{
		Strategy: BackoffStrategyExponentialWithJitter,
		MaxTries: 10,
		Period:   0,
	}
	if err := rpZero.Backoff(canceledCtx, 1); err == nil || err.Error() != "context has been cancelled" {
		t.Errorf("Backoff() with canceled context got %v, want 'context has been cancelled'", err)
	}
}
