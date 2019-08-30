package x

import (
	"context"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// Message is the interface to a transport-specific message containing an event.
//
// There are 3 reliable qualities of service for messages:
//
// 0/at-most-once/unreliable: messages can be dropped silently.
//
// 1/at-least-once: messages are not dropped without signaling an error
// at the sender but may be duplicated.
//
// 2/exactly-once: messages are never dropped (without error) or
// duplicated, as long as both ends maintain some transport-specific
// state.
//
// The Message interface supports QoS 0 and 1, the ExactlyOnceMessage inteface
// supports QoS 2
//
type Message interface {
	// Event decodes and returns the contained Event.
	Event() (cloudevents.Event, error)

	// Structured optionally returns an encoded structured event if it
	// is efficient to do so, or ("", nil) if not.
	//
	// Enables Senders to optimize the case where structured events are
	// passed from transport to transport without being decoded and
	// re-encoded.
	//
	// Transport Message and Sender implementations are not required to
	// implement the optimization, they can ignore it.
	Structured() (encodingMediaType string, encodedEvent []byte)

	// Finish *must* be called when message from a Receiver can be
	// forgotten by the receiver.
	//
	// A QoS 1 sender forwarding messages should not call Finish()
	// until it gets an acknowledgment of receipt.
	//
	// A non-nil error indicates an error in sending or processing.
	Finish(error)
}

// ExactlyOnceMessage is implemented by incoming transport messages
// that support QoS 2.  Only transports that support QoS 2 need to
// implement or use this interface.
type ExactlyOnceMessage interface {
	Message

	// Received is called by a Sender when it gets acknowledgment of receipt
	// (e.g. AMQP ACCEPT or MQTT PUBREC)
	//
	// The sender passes a finish() function that the original receiver
	// must call when it get's the ack-of-the-ack (e.g. AMQP SETTLE, MQTT
	// PUBCOMP)
	//
	// If sending fails, the sender must call Finish(err) with a non-nil
	// error instead of Received. ExactlyOnceMessage implementations
	// must also be prepared to handle Finish(nil) if the sender does
	// not support QoS 3.
	Received(finish func(error))
}

// EventMessage wraps a local cloudevents.Event as a Message.
type EventMessage cloudevents.Event

func (m EventMessage) Event() (cloudevents.Event, error) { return cloudevents.Event(m), nil }
func (_ EventMessage) Structured() (string, []byte)      { return "", nil }
func (_ EventMessage) Finish(error)                      {}

var _ Message = EventMessage{} // Test it conforms to the interface

// Receiver is the receiving half of a transport.
type Receiver interface {
	Receive(ctx context.Context) (Message, error)
}

// Sender is the sending half of a transport.
type Sender interface {
	Send(ctx context.Context, m Message) error
}

// Transport implements the transport.Transport interface using a
// Sender and Receiver.
type Transport struct {
	Sender   Sender
	Receiver Receiver
	handler  transport.Receiver
}

var _ transport.Transport = (*Transport)(nil) // Test it conforms to the interface

func NewTransport(s Sender, r Receiver) *Transport {
	return &Transport{Sender: s, Receiver: r}
}

func (t *Transport) Send(ctx context.Context, e cloudevents.Event) (context.Context, *cloudevents.Event, error) {
	return ctx, nil, t.Sender.Send(ctx, EventMessage(e))
}

func (t *Transport) SetReceiver(r transport.Receiver) { t.handler = r }

func (t *Transport) StartReceiver(ctx context.Context) error {
	for {
		if m, err := t.Receiver.Receive(ctx); err != nil {
			return err
		} else if e, err := m.Event(); err != nil {
			m.Finish(err)
			return err
		} else if err := t.handler.Receive(ctx, e, nil); err != nil {
			m.Finish(err)
			return err
		} else {
			m.Finish(nil)
		}
	}
}

func (t *Transport) SetConverter(transport.Converter) {
	// TODO(alanconway) Can we separate Converter from the base transport interface?
}

func (t *Transport) HasConverter() bool {
	return false
}
