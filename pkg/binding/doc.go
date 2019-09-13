/*

Package binding provides interfaces for transport bindings.
Binding implementations are under ../bindings.

Intermediary applications that forward events between different
transport bindings may also need to use these intefaces. Normal
clients can use the cloudevents/client API.

A transport binding implements Message, Sender, Receiver interfaces.
An intermediary uses those interfaces to transfer event messages.  A
binding should also provide functions to convert between
cloudevents.Event and the binding's native message types. There are no
interfaces as the native message type will vary by binding.

A Message is an abstract container for a cloudevents.Event.

A Receiver returns Message implementations based on its native message
format. A Sender must be able to encode and send any Message, but may
provide optimized handling for StructMessage, or for its own native
Message implementations.

Transports that support reliable delivery can implement and use
Message.Finish() and ExactlyOnceMessage to forward acknowledgement
between sender and receiver for QoS level 0, 1 or 2. The effective QoS
of a sender/receiver pair will be the lower of the two.

FIXME(alanconway) add a generic encoder interface or function signature.  The
Sender implicitly uses an encoder, so it's not needed for normal use.  It is
useful for testing and for users with inside knowledge of the implementation,
they may want to use the encoder but send messages by some other means than the
Sender.

*/
package binding
