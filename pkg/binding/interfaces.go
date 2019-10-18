package binding

import (
	"context"

	ce "github.com/cloudevents/sdk-go/pkg/cloudevents"
)

// Message is the interface to a binding-specific message containing an event.
//
// Reliable Delivery
//
// There are 3 reliable qualities of service for messages:
//
// 0/at-most-once/unreliable: messages can be dropped silently.
//
// 1/at-least-once: messages are not dropped without signaling an error
// to the sender, but they may be duplicated in the event of a re-send.
//
// 2/exactly-once: messages are never dropped (without error) or
// duplicated, as long as both sending and receiving ends maintain
// some binding-specific delivery state. Whether this is persisted
// depends on the configuration of the binding implementations.
//
// The Message interface supports QoS 0 and 1, the ExactlyOnceMessage interface
// supports QoS 2
//
type Message interface {
	// Event decodes and returns the contained Event.
	Event() (ce.Event, error)

	// Structured optionally returns a structured event encoding and its
	// format type, if the message contains such an encoding. Returns
	// ("", nil) if not.
	//
	// This allows Senders to avoid re-encoding messages that are
	// already in suitable structured form.
	//
	Structured() (formatMediaType string, encodedEvent []byte)

	// Finish *must* be called when message from a Receiver can be forgotten by
	// the receiver. Sender.Send() calls Finish() when the message is sent.  A QoS
	// 1 sender should not call Finish() until it gets an acknowledgment of
	// receipt on the underlying transport.  For QoS 2 see ExactlyOnceMessage.
	//
	// Passing a non-nil err indicates sending or processing failed.
	// A non-nil return indicates that the message was not accepted
	// by the receivers peer.
	Finish(error) error
}

// ExactlyOnceMessage is implemented by received Messages
// that support QoS 2.  Only transports that support QoS 2 need to
// implement or use this interface.
type ExactlyOnceMessage interface {
	Message

	// Received is called by a forwarding QoS2 Sender when it gets
	// acknowledgment of receipt (e.g. AMQP 'accept' or MQTT PUBREC)
	//
	// The receiver must call settle(nil) when it get's the ack-of-ack
	// (e.g. AMQP 'settle' or MQTT PUBCOMP) or settle(err) if the
	// transfer fails.
	//
	// Finally the Sender calls Finish() to indicate the message can be
	// discarded.
	//
	// If sending fails, or if the sender does not support QoS 2, then
	// Finish() may be called without any call to Received()
	Received(settle func(error))
}

// Receiver receives messages.
type Receiver interface {
	// Receive blocks till a message is received or ctx expires.
	//
	// A non-nil error means the receiver is closed.
	// io.EOF means it closed cleanly, any other value indicates an error.
	Receive(ctx context.Context) (Message, error)
}

// Sender sends messages.
type Sender interface {
	// Send a message.
	//
	// Send returns when the "outbound" message has been sent. The Sender may
	// still be expecting acknowledgment or holding other state for the message.
	//
	// m.Finish() is called when sending is finished: expected acknowledgments (or
	// errors) have been received, the Sender is no longer holding any state for
	// the message. m.Finish() may be called during or after Send().
	//
	// To support optimized forwading of structured-mode messages, Send()
	// should use the encoding returned by m.Structured() if there is one.
	// Otherwise m.Event() can be encoded as per the binding's rules.
	Send(ctx context.Context, m Message) error
}

// Requester sends a message and receives a response
//
// Optional interface that may be implemented by protocols that support
// request/response correlation.
type Requester interface {
	// Request sends m like Sender.Send() but also arranges to receive a response.
	// The returned Receiver is used to receive the response.
	Request(ctx context.Context, m Message) (Receiver, error)
}

// Closer is the common interface for things that can be closed
type Closer interface {
	Close(ctx context.Context) error
}

// ReceiveCloser is a Receiver that can be closed.
type ReceiveCloser interface {
	Receiver
	Closer
}

// SendCloser is a Sender that can be closed.
type SendCloser interface {
	Sender
	Closer
}

// Encoder encodes events as messages.
type Encoder interface {
	Encode(ce.Event) (Message, error)
}
