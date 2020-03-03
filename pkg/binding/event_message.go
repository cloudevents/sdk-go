package binding

import (
	"bytes"
	"context"

	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/event"
)

// EventMessage type-converts a event.Event object to implement Message.
// This allows local event.Event objects to be sent directly via Sender.Send()
//     s.Send(ctx, binding.EventMessage(e))
type EventMessage event.Event

func (m EventMessage) GetParent() Message {
	return nil
}

func (m EventMessage) Encoding() Encoding {
	return EncodingEvent
}

func (m EventMessage) Structured(ctx context.Context, builder StructuredEncoder) error {
	// TODO here only json is supported, should we support other message encodings?
	b, err := format.JSON.Marshal(event.Event(m))
	if err != nil {
		return err
	}
	return builder.SetStructuredEvent(ctx, format.JSON, bytes.NewReader(b))
}

func (m EventMessage) Binary(ctx context.Context, b BinaryEncoder) (err error) {
	err = b.Start(ctx)
	if err != nil {
		return err
	}
	err = EventContextToBinaryEncoder(m.Context, b)
	if err != nil {
		return err
	}
	// Pass the body
	body, err := (*event.Event)(&m).DataBytes()
	if err != nil {
		return err
	}
	if len(body) > 0 {
		err = b.SetData(bytes.NewReader(body))
		if err != nil {
			return err
		}
	}
	return b.End()
}

func (EventMessage) Finish(error) error { return nil }

func (m *EventMessage) SetEvent(e event.Event) error {
	*m = EventMessage(e)
	return nil
}

var _ Message = (*EventMessage)(nil) // Test it conforms to the interface

func EventContextToBinaryEncoder(c event.EventContext, b BinaryEncoder) (err error) {
	// Pass all attributes
	var sv spec.Version
	sv, err = spec.VS.Version(c.GetSpecVersion())
	if err != nil {
		return err
	}
	for _, a := range sv.Attributes() {
		value := a.Get(c)
		if value != nil {
			err = b.SetAttribute(a, value)
		}
		if err != nil {
			return err
		}
	}
	// Pass all extensions
	for k, v := range c.GetExtensions() {
		err = b.SetExtension(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
