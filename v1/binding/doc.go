/*

Package binding defines interfaces for protocol bindings.

NOTE: Most applications that emit or consume events should use the ../client
package, which provides a simpler API to the underlying binding.

The interfaces in this package provide extra encoding and protocol information
to allow efficient forwarding and end-to-end reliable delivery between a
Receiver and a Sender belonging to different bindings. This is useful for
intermediary applications that route or forward events, but not necessary for
most "endpoint" applications that emit or consume events.

Protocol Bindings

A protocol binding implements at least Message, Sender and Receiver, and usually
Encoder.

Receiver: receives protocol messages and wraps them to implement the Message interface.

Message: interface that defines the visitors for an encoded event in structured mode,
binary mode or event mode. A method is provided to read the Encoding of the message

Sender: converts arbitrary Message implementations to a protocol-specific form
and sends them. A protocol Sender should preserve the spec-version and
structured/binary mode of sent messages as far as possible. This package
provides generic Sender wrappers to pre-process messages into a specific
spec-version or structured/binary mode when the user requires that.

Message and ExactlyOnceMessage provide methods to allow acknowledgments to
propagate when a reliable messages is forwarded from a Receiver to a Sender.
QoS 0 (unreliable), 1 (at-least-once) and 2 (exactly-once) are supported.


Intermediaries

Intermediaries can forward Messages from a Receiver to a Sender without
knowledge of the underlying protocols. The Message interface allows structured
messages to be forwarded without decoding and re-encoding. It also allows any
Message to be fully decoded and examined as needed.

*/
package binding
