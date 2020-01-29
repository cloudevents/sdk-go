package event

import (
	"bytes"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

var specs = spec.New()

// EventMessage type-converts a cloudevents.Event object to implement Message.
// This allows local cloudevents.Event objects to be sent directly via Sender.Send()
//     s.Send(ctx, binding.EventMessage(e))
type EventMessage ce.Event

func (m EventMessage) Event(builder binding.EventEncoder) error {
	return builder.SetEvent(ce.Event(m))
}

func (m EventMessage) Structured(binding.StructuredEncoder) error {
	return binding.ErrNotStructured
}

func (m EventMessage) Binary(b binding.BinaryEncoder) (err error) {
	err = EventContextToBinaryEncoder(m.Context, b)
	if err != nil {
		return err
	}
	// Pass the body
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

var _ binding.Message = (*EventMessage)(nil) // Test it conforms to the interface
var _ binding.EventEncoder = (*EventMessage)(nil)

func EventContextToBinaryEncoder(c cloudevents.EventContext, b binding.BinaryEncoder) (err error) {
	// Pass all attributes
	var sv spec.Version
	sv, err = specs.Version(c.GetSpecVersion())
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
