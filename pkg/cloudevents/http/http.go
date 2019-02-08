package http

import (
	"net/http"
	"reflect"

	"github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"
)

// HTTPMarshaller an interface with methods for creating CloudEvents
type HTTPMarshaller interface {
	FromRequest(req *http.Request) (canonical.Event, error)
	ToRequest(req *http.Request, event canonical.Event) error
}

// HTTPCloudEventConverter an interface for defining converters that can read/write CloudEvents from HTTP requests
type HTTPCloudEventConverter interface {
	CanRead(t reflect.Type, mediaType string) bool
	CanWrite(t reflect.Type, mediaType string) bool
	Read(t reflect.Type, req *http.Request) (canonical.Event, error)
	Write(t reflect.Type, req *http.Request, event canonical.Event) error
}
