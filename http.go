package cloudevents

import (
	"net/http"

	"github.com/dispatchframework/cloudevents-go-sdk/v01"
)

// FromHTTPRequest parses a CloudEvent from any known encoding.
func FromHTTPRequest(req *http.Request) (Event, error) {
	// TODO: this should check the version of incoming CloudVersion header and create an appropriate event structure.
	e := &v01.Event{}
	err := e.FromHTTPRequest(req)
	return e, err

}
