/*

Package binding is for implementing transport bindings and
intermediaries like importers, brokers or channels that forward
messages between bindings.

A transport binding implements Message, Sender and Receiver interfaces.
An intermediary uses those interfaces to transfer event messages.

A Message is an abstract container for an Event. It provides
additional methods for efficient forwarding of structured events
and reliable delivery when the underlying transports support it.

A Receiver can return instances of its own Message implementations. A
Sender must be able to send any implementation of the Message
interface, but it may provide optimized handling for its own Message
implementations.

For transports that support a reliable delivery QoS, the Message
interface allows acknowledgment between sender and receiver for QoS
level 0, 1 or 2. The effective QoS is the lowest provided by sender or
receiver.

*/
package binding
