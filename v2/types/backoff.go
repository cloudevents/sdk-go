package types

import "time"

// Backoff holds parameters applied to retries
type Backoff struct {

	// Retry is the number of times to retry request before giving up
	Retry int

	// Delay is the delay before retrying.
	// For linear policy, backoff delay is the time interval between retries.
	Delay time.Duration
}
