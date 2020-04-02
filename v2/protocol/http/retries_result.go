package http

import (
	"fmt"

	"github.com/cloudevents/sdk-go/v2/protocol"
)

// NewRetriesResult returns a http RetriesResult that should be used as
// a transport.Result without retries
func NewRetriesResult(result protocol.Result, retries int, results []protocol.Result) protocol.Result {
	return &RetriesResult{
		Result:  result,
		Retries: retries,
		Results: results,
	}
}

// Result wraps the fields required to make adjustments for http Responses.
type RetriesResult struct {
	// The last result
	protocol.Result

	// Retries is the number of times the request was tried
	Retries int

	//
	Results []protocol.Result
}

// make sure RetriesResult implements error.
var _ error = (*RetriesResult)(nil)

// Is returns if the target error is a RetriesResult type checking target.
func (e *RetriesResult) Is(target error) bool {
	return protocol.ResultIs(e.Result, target)
}

// Error returns the string that is formed by using the format string with the
// provided args.
func (e *RetriesResult) Error() string {
	if e.Retries == 0 {
		return e.Result.Error()
	}
	return fmt.Sprintf("%s (%dx)", e.Result.Error(), e.Retries)
}
