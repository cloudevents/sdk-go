package transport

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"net/http"
)

// Transport is the interface for transport sender to send the converted Message
// over the underlying transport.
type Sender interface {
	Send(cloudevents.Event, *http.Request) (*http.Response, error) // TODO: these leaks the http request.
}

// Receiver TODO not sure yet.
type Receiver interface {
	Receive(cloudevents.Event)
}
