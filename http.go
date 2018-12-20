package cloudevents

import (
	"net/http"
	"reflect"
)

// HTTPMarshaller an interface with methods for creating CloudEvents
type HTTPMarshaller interface {
	FromRequest(req *http.Request, event interface{}) error
	ToRequest(req *http.Request, event interface{}) error
}

// HTTPCloudEventConverter an interface for defining converters that can read/write CloudEvents from HTTP requests
type HTTPCloudEventConverter interface {
	CanRead(t reflect.Type, mediaType string) bool
	CanWrite(t reflect.Type, mediaType string) bool
	Read(req *http.Request, event interface{}) error
	Write(req *http.Request, event interface{}) error
}

type BinaryMarshaler interface {
	MarshalBinary(req *http.Request) error
}

type BinaryUnmarshaler interface {
	UnmarshalBinary(req *http.Request) error
}

type HasContentType interface {
	GetContentType() string
}
