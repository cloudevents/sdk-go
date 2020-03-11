# CloudEvents Golang SDK v2

Version 2 of this SDK takes lessons learned from the original effort and updates
the API into how we have seen v1 integrate into systems.

## Terms

- [Event](https://github.com/cloudevents/spec/blob/master/spec.md#event), an
  Event is the conical form of the attributes and payload of the occurrence.
- [Message](https://github.com/cloudevents/spec/blob/master/spec.md#message), a
  Message is the encoded form of an Event for a given encoding and protocol.
- [Protocol](https://github.com/cloudevents/spec/blob/master/spec.md#protocol),
  a Protocol is the over-the-wire format that Messages are sent.
- [Protocol Binding](https://github.com/cloudevents/spec/blob/master/spec.md#protocol-binding),
  a Protocol Binding defines how Events are mapped into Messages for the given
  Protocol.
- Client, a client contains the logic of interacting with a Protocol to enable
  interactions with Events. Clients also provide protocol agnostic features that
  can be applied to events, such as extensions.
- Extensions, an extension is anything that extends the base requirements from
  the CloudEvents spec. There are several
  [CloudEvents supported extensions](https://github.com/cloudevents/spec/tree/master/extensions).
- Message Writer, the logic required to take in a Message and write out to a
  given Protocol (request, message).

## Investment Level

The amount of the SDK adopters would like to use is up to the adopter. We
support the following:

- Resource Level, an Event can be used directly, and can be marshaled in and and
  out of JSON.
- Protocol Level, a Protocol can be used directly, with facilities to aid in
  converting an Event into a Message by using Message Readers and Writers for
  that protocol.
- Client Level, a Protocol can be selected and Events can be directly sent and
  received without requiring interactions with Message objects.

## Personas

- [Producer](https://github.com/cloudevents/spec/blob/master/spec.md#producer),
  the "producer" is a specific instance, process or device that creates the data
  structure describing the CloudEvent.
- [Consumer](https://github.com/cloudevents/spec/blob/master/spec.md#consumer),
  a "consumer" receives the event and acts upon it. It uses the context and data
  to execute some logic, which might lead to the occurrence of new events.
- [Intermediary](https://github.com/cloudevents/spec/blob/master/spec.md#intermediary),
  An "intermediary" receives a message containing an event for the purpose of
  forwarding it to the next receiver, which might be another intermediary or a
  Consumer. A typical task for an intermediary is to route the event to
  receivers based on the information in the Context.

## Interaction Models

The SDK enables the following interaction models.

### Sender

Sender, when a Producer is creating new events.

![sender](./images/sender.svg "Sender")

### Receiver

Receiver, when a Consumer is accepting events.

![receiver](./images/receiver.svg "Receiver")

### Forwarder

Forwarder, when a Intermediary accepts an event only after it has successfully
continued the message to one or more Consumers.

![forwarder](./images/forwarder.svg "Forwarder")

### Mutator

Mutator, when a Producer or Intermediary blocks on a response from a Consumer,
replacing the original Event.

![mutator](./images/mutator.svg "Mutator")

---

# For Integrators

To leverage the SDK, the following interfaces need to be considered:

Event -> Message -> bits

bits -> Message -> Event

bit -> Message -> bits

## Interfaces (Current pre-v2)

```
Client --> Transport (via TransportBinding) -> Protocol (implements Sender, Receiver)* -> Write<DataStructure>
```

`Write<DataStructure>` functions read a `binding.Message` and write what is
found into the [3pl][3pl] data structure. For example, `nats.WriteMsg` writes
the message into a `nats.Msg`, or `http.WriteRequest` writes message into a
`http.Request`.

Protocol is the thinnest wrapper for the [3pl][3pl] to implement Sender and
Receiver.

\*Protocol can optionally implement Requester and Responder(not yet defined,
receiver with reply). Protocol does not want to manage the underlying resources
that enable the external communication. For example: http.Protocol implements a
ServeHTTP method which can be leveraged by another SDK component:
http.Transport, or used directly in custom integrations where more control is
required.

Transport is legacy. It connects the Client to the Protocol and also manages a
lot of the internal [3pl][3pl] setup. This is too much overhead for the SDK.

Client gives a simple API for sending and receiving events as event.Event
objects.

[3pl]: 3pl: 3rd party lib (nats.io or net/http)
