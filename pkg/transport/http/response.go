package http

import (
	"errors"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/event"
)

// NewResponse returns a fully populated http Response that should be used as
// a event.Response.
func NewResponse(status int, messageFmt string, args ...interface{}) event.Response {
	return &Response{
		Status: status,
		Format: messageFmt,
		Args:   args,
	}
}

// Response wraps the fields required to make adjustments for http Responses.
type Response struct {
	Status int
	Format string
	Args   []interface{}
}

// make sure Response implements error.
var _ error = (*Response)(nil)

// Is returns if the target error is a Response type checking target.
func (e *Response) Is(target error) bool {
	if _, ok := target.(*Response); ok {
		return true
	}
	// Allow for wrapped errors.
	err := fmt.Errorf(e.Format, e.Args...)
	return errors.Is(err, target)
}

// Error returns the string that is formed by using the format string with the
// provided args.
func (e *Response) Error() string {
	return fmt.Sprintf(e.Format, e.Args...)
}
