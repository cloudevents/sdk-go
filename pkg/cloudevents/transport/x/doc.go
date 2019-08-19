/*

Package transport/x shows proposed modifications to package transport.

They are here for discussion purposes, anything that we can agree on
should be merged into the transport package and may involve
incompatible changes.

Goals

Support for intermediaries that receive and send events via different
transports, for example adapters, importers, brokers or channels.

Support for reliable messaging qualities of service between different
transports

Provide the option of blocking Receive vs. callback SetReceiver.
There are many situations where the blocking style is easier to
implement and work with.

Overview of changes

Sender/Receiver interfaces that can be implemented and used separately.
Implementations can easily be combined into a single Transport that does both
(e.g. for use by the client package)

Message.Event() allows conversion from any message to the common
in-memory Event representation, without knowing its transport/codec of
origin. Explict use of Codecs is not required. (Transports should
still provide a way for users to build/interpret transport messages if they
want to bypass our transport APIs, but that's a separate issue.)

Message.Structured() optionally allows a structured encoding to be
forwarded without decoding. This is a special case, but an important
one.

Reliable messaging: Transport implementation must provide the actual
reliability (resends, storing delivery state etc.) the Message
interface allows hand-over responsibility between a Receiver and a
Sender on different transports so we don't create a "hole" in the
guarantee. The overall guarantee is the weaker of the sender and
receiver.

No reliability examples yet, wanted to get some feedback before going further.
*/
package x

// FIXME(alanconway) review - check for "2" suffix
// Finish comments.
