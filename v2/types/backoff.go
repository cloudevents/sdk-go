package types

import "time"

type BackoffStrategy string

const (
	BackoffStrategyLinear      = "linear"
	BackoffStrategyExponential = "exponential"
)

// Backoff holds parameters applied to retries
type Backoff struct {
	// Backoff strategy
	Strategy BackoffStrategy

	// MaxTries is the maximum number of times to retry request before giving up
	MaxTries int

	// Period is
	// - for linear strategy: the delay interval between retries
	// - for exponential strategy: the factor applied after each retry
	Period time.Duration
}
