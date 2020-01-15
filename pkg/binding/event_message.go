package binding

import (
	"bytes"

	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

var specs = spec.New()

// EventMessage type-converts a cloudevents.Event object to implement Message.
// This allows local cloudevents.Event objects to be sent directly via Sender.Send()
//     s.Send(ctx, binding.EventMessage(e))
type EventMessage ce.Event

func (m EventMessage) Event(builder EventEncoder) error {
	return builder.SetEvent(ce.Event(m))
}

func (m EventMessage) Structured(StructuredEncoder) error {
	return ErrNotStructured
}

func (m EventMessage) Binary(b BinaryEncoder) (err error) {
	c := m.Context
	var sv spec.Version
	sv, err = specs.Version(c.GetSpecVersion())
	if err != nil {
		return err
	}
	set := func(k spec.Kind, v interface{}) {
		if err == nil {
			attr := sv.AttributeFromKind(k)
			if attr != nil {
				err = b.SetAttribute(attr, v)
			}
		}
	}
	set(spec.SpecVersion, c.GetSpecVersion())
	set(spec.Type, c.GetType())
	set(spec.Source, c.GetSource())
	set(spec.Subject, c.GetSubject())
	set(spec.ID, c.GetID())
	set(spec.Time, c.GetTime())
	set(spec.DataSchema, c.GetDataSchema())
	set(spec.DataContentType, c.GetDataContentType())

	for k, v := range c.GetExtensions() {
		if err == nil {
			err = b.SetExtension(k, v)
		}
	}
	if err != nil {
		return err
	}
	body, err := (*ce.Event)(&m).DataBytes()
	if err == nil && len(body) > 0 {
		return b.SetData(bytes.NewReader(body))
	}
	return err
}

func (EventMessage) Finish(error) error { return nil }

func (m *EventMessage) SetEvent(e ce.Event) error {
	*m = EventMessage(e)
	return nil
}

var _ Message = (*EventMessage)(nil) // Test it conforms to the interface
