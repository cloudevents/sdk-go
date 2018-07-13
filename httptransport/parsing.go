package httptransport

import (
	"errors"
	"net/http"

	"github.com/dispatchframework/cloudevents-go-sdk"
)

// TODO: since transport is versioned, should this be versioned together with event implementation?

// Format type wraps supported modes of formatting CloudEvent as HTTP request.
// Currently, only binary mode and structured mode with JSON encoding are supported.
type Format int

const (
	// FormatBinary corresponds to Binary mode in CloudEvents HTTP transport binding.
	// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/http-transport-binding.md#31-binary-content-mode
	FormatBinary Format = iota
	// FormatJSON corresponds to Structured mode using JSON encoding.
	// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/http-transport-binding.md#32-structured-content-mode
	FormatJSON
)

// EventFromRequest parses the http request and returns a CloudEvent.
// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/http-transport-binding.md
func EventFromRequest(req *http.Request) (cloudevents.CloudEvent, error) {
	return nil, errors.New("not implemented")

}

// EventToRequest creates a http request using a format specified in CloudEvents HTTP transport binding.
// https://github.com/cloudevents/spec/blob/a12b6b618916c89bfa5595fc76732f07f89219b5/http-transport-binding.md#31-binary-content-mode
func EventToRequest(event cloudevents.CloudEvent, req *http.Request, format Format) error {
	return errors.New("not implemented")

}
