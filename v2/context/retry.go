/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

type BackoffStrategy string

const (
	BackoffStrategyNone                  = "none"
	BackoffStrategyConstant              = "constant"
	BackoffStrategyLinear                = "linear"
	BackoffStrategyExponential           = "exponential"
	BackoffStrategyExponentialWithJitter = "exponential-jitter"
)

var DefaultRetryParams = RetryParams{Strategy: BackoffStrategyNone}

// RetryParams holds parameters applied to retries
type RetryParams struct {
	// Strategy is the backoff strategy to applies between retries
	Strategy BackoffStrategy

	// MaxTries is the maximum number of times to retry request before giving up
	MaxTries int

	// Period is
	// - for none strategy: no delay
	// - for constant strategy: the delay interval between retries
	// - for linear strategy: interval between retries = Period * retries
	// - for exponential strategy: interval between retries = Period * retries^2
	// - for exponential-jitter strategy: random interval between 0 and Period * 2^retries
	Period time.Duration
}

// BackoffFor tries will return the time duration that should be used for this
// current try count.
// `tries` is assumed to be the number of times the caller has already retried.
func (r *RetryParams) BackoffFor(tries int) time.Duration {
	switch r.Strategy {
	case BackoffStrategyConstant:
		return r.Period
	case BackoffStrategyLinear:
		return r.Period * time.Duration(tries)
	case BackoffStrategyExponential:
		exp := math.Exp2(float64(tries))
		return r.Period * time.Duration(exp)
	case BackoffStrategyExponentialWithJitter:
		if r.Period <= 0 {
			return 0
		}
		ceiling := float64(r.Period) * math.Exp2(float64(tries))
		if ceiling >= float64(math.MaxInt64) {
			return time.Duration(rand.Int63n(math.MaxInt64))
		}
		maxCeil := int64(ceiling)
		if maxCeil <= 0 {
			return 0
		}
		return time.Duration(rand.Int63n(maxCeil))
	case BackoffStrategyNone:
		fallthrough // default
	default:
		return r.Period
	}
}

// Backoff is a blocking call to wait for the correct amount of time for the retry.
// `tries` is assumed to be the number of times the caller has already retried.
func (r *RetryParams) Backoff(ctx context.Context, tries int) error {
	if tries > r.MaxTries {
		return errors.New("too many retries")
	}
	d := r.BackoffFor(tries)
	if d <= 0 {
		select {
		case <-ctx.Done():
			return errors.New("context has been cancelled")
		default:
			return nil
		}
	}
	ticker := time.NewTicker(d)
	select {
	case <-ctx.Done():
		ticker.Stop()
		return errors.New("context has been cancelled")
	case <-ticker.C:
		ticker.Stop()
	}
	return nil
}
