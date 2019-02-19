package cloudevents

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/datacodec"
)

// Event represents the canonical representation of a CloudEvent.
type Event struct {
	Context EventContext
	Data    interface{}
}

func (e Event) DataAs(data interface{}) error {
	return datacodec.Decode(e.Context.GetDataContentType(), e.Data, data)
}
