package receiver

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"sync"
)

// TypedReceiver is a simple router to invoke a Receiver when the incoming CloudEvent matches Type _exactly_.
type TypedReceiver struct {
	trigger map[string]DynamicReceiver
	once    sync.Once
}

// Add registers fn for eventType. If eventType is already set, this overwrites it.
func (r *TypedReceiver) Add(eventType string, fn interface{}) error {
	var dr DynamicReceiver
	var err error
	if dr, err = NewDynamicReceiver(fn); err != nil {
		return err
	}

	r.once.Do(func() {
		if r.trigger == nil {
			r.trigger = make(map[string]DynamicReceiver)
		}
	})

	r.trigger[eventType] = dr
	return nil
}

// Remove unregisters eventType from this TypedReceiver.
func (r *TypedReceiver) Remove(eventType string) {
	if r.trigger != nil {
		delete(r.trigger, eventType)
	}
}

// Receive implements
func (r *TypedReceiver) Receive(ctx context.Context, event cloudevents.Event, resp *cloudevents.EventResponse) error {
	if fn, ok := r.trigger[event.Type()]; ok {
		return fn.Invoke(ctx, event, resp)
	}
	return nil
}
