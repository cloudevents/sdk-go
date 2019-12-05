package binding

import ce "github.com/cloudevents/sdk-go/pkg/cloudevents"

// EventMessage type-converts a cloudevents.Event object to implement Message.
// This allows local cloudevents.Event objects to be sent directly via Sender.Send()
//     s.Send(ctx, binding.EventMessage(e))
type EventMessage ce.Event

func (m EventMessage) Event(builder EventMessageBuilder) error   { return builder.Encode(ce.Event(m)) }
func (m EventMessage) Structured(StructuredMessageBuilder) error { return ErrNotStructured }
func (m EventMessage) Binary(BinaryMessageBuilder) error         { return ErrNotBinary }
func (EventMessage) Finish(error) error                          { return nil }

var _ Message = (*EventMessage)(nil) // Test it conforms to the interface
