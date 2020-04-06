package context

import "time"

type BackoffStrategy string

const (
	BackoffStrategyNone   = "none"
	BackoffStrategyLinear = "linear"

	// TODO
	// BackoffStrategyExponential = "exponential"
)

var DefaultRetryParams = RetryParams{Strategy: BackoffStrategyNone}

// RetryParams holds parameters applied to retries
type RetryParams struct {
	// Strategy is the backoff strategy to applies between retries
	Strategy BackoffStrategy

	// MaxTries is the maximum number of times to retry request before giving up
	MaxTries int

	// Period is
	// - for linear strategy: the delay interval between retries
	// - for exponential strategy: the factor applied after each retry
	Period time.Duration
}
