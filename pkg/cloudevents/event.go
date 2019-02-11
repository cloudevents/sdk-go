package cloudevents

import (
	"encoding/json"
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/context"
)

// Event represents the canonical representation of a CloudEvent.
type Event struct {
	Context context.EventContext
	Data    interface{}
}

func (e Event) DataAs(data interface{}) error {
	switch e.Context.DataContentType() {
	case "application/json":
		if b, ok := e.Data.([]byte); !ok {
			data = e.Data
		} else if err := json.Unmarshal(b, data); err != nil {
			return fmt.Errorf("found json, but failed to unmarshal: %s", err.Error())
		}
	default:
		return fmt.Errorf("not implemented")
	}
	return nil
}
