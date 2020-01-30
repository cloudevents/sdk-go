package binding

import (
	"context"
	"errors"
	"io"

	"github.com/cloudevents/sdk-go/pkg/binding/format"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
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
// The Structured and Binary methods provide optional optimized transfer of an event
// to a Sender, they may not be implemented by all Message instances. A Sender should
// try each method of interest and fall back to Event(EventEncoder) if none are supported.
//
type Message interface {
	// Structured transfers a structured-mode event to a StructuredEncoder.
	// Returns ErrNotStructured if message is not in structured mode.
	//
	// Returns a different err if something wrong happened while trying to read the structured event
	// In this case, the caller must Finish the message with appropriate error
	//
	// This allows Senders to avoid re-encoding messages that are
	// already in suitable structured form.
	Structured(StructuredEncoder) error

	// Binary transfers a binary-mode event to an BinaryEncoder.
	// Returns ErrNotBinary if message is not in binary mode.
	//
	// Returns a different err if something wrong happened while trying to read the binary event
	// In this case, the caller must Finish the message with appropriate error
	//
	// Allows Senders to forward a binary message without allocating an
	// intermediate Event.
	Binary(BinaryEncoder) error

	// Event transfers an event to an EventEncoder.
	//
	// A message implementation should always provide a conversion to Event data structure.
	// The implementor can use binding.ToEvent for a straightforward implementation, starting
	// from binary or structured representation
	//
	// Returns an err if something wrong happened while trying to read the event
	// In this case, the caller must Finish the message with appropriate error
	//
	// Useful when the Sender can't implement a direct binary/structured to binary/structured conversion,
	// So the intermediate Event representation is required
	Event(EventEncoder) error

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

// ErrNotStructured returned by Message.Structured for non-structured messages.
var ErrNotStructured = errors.New("message is not in structured mode")

// ErrNotBinary returned by Message.Binary for non-binary messages.
var ErrNotBinary = errors.New("message is not in binary mode")

// StructuredEncoder should generate a new representation of the event starting from a structured message.
//
// Protocols that supports structured encoding should implement this interface to implement direct
// structured -> structured transfer.
type StructuredEncoder interface {
	// Event receives an io.Reader for the whole event.
	SetStructuredEvent(format format.Format, event io.Reader) error
}

// BinaryEncoder should generate a new representation of the event starting from a binary message.
//
// Protocols that supports binary encoding should implement this interface to implement direct
// binary -> binary transfer.
type BinaryEncoder interface {
	// SetData receives an io.Reader for the data attribute.
	// io.Reader could be empty, meaning that message payload is empty
	SetData(data io.Reader) error

	// Set a standard attribute.
	//
	// The value can either be the correct golang type for the attribute, or a canonical
	// string encoding. See package cloudevents/types
	SetAttribute(attribute spec.Attribute, value interface{}) error

	// Set an extension attribute.
	//
	// The value can either be the correct golang type for the attribute, or a canonical
	// string encoding. See package cloudevents/types
	SetExtension(name string, value interface{}) error
}

// EventEncoder should generate a new representation of the event starting from an event message.
//
// Every protocol must implement this interface. If a protocol supports both structured and binary encoding,
// two EventEncoder implementations could be provided
type EventEncoder interface {
	SetEvent(ce.Event) error
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
