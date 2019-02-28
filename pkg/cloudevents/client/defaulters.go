package client

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/google/uuid"
)

type EventDefaulter func(event cloudevents.Event) cloudevents.Event

func DefaultIDToUUIDIfNotSet(event cloudevents.Event) cloudevents.Event {
	if event.Context != nil {
		switch event.Context.GetSpecVersion() {
		case cloudevents.CloudEventsVersionV01:
			ec := event.Context.AsV01()
			if ec.EventID == "" {
				ec.EventID = uuid.New().String()
				event.Context = ec
			}
		case cloudevents.CloudEventsVersionV02:
			ec := event.Context.AsV02()
			if ec.ID == "" {
				ec.ID = uuid.New().String()
				event.Context = ec
			}
		case cloudevents.CloudEventsVersionV03:
			ec := event.Context.AsV03()
			if ec.ID == "" {
				ec.ID = uuid.New().String()
				event.Context = ec
			}
		}
	}
	return event
}
