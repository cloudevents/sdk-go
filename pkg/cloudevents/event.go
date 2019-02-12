package cloudevents

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents/context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/datacodec"
)

// Event represents the canonical representation of a CloudEvent.
type Event struct {
	Context context.EventContext
	Data    interface{}
}

func (e Event) DataAs(data interface{}) error {
	return datacodec.Decode(e.Context.DataContentType(), e.Data, data)
}
