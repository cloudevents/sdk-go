---
title: Protocol Bindings
nav_order: 4
---

# Protocol Binding implementations
{: .no_toc }

1. TOC
{:toc}

## Overview

Every supported protocol binding implements the logic to read an incoming CloudEvent message implementing 
the [`Message` interface](https://github.com/cloudevents/sdk-go/tree/master/v2/binding/message.go) and the logic to write out a CloudEvent message
providing `Write<DataStructure>` functions.
Then a bunch of interfaces are implemented to allow the user to interact, through the `Client`, with the given protocol.

## `Message` interface

`Message` is the interface to a binding-specific message containing an event. 
This interface abstracts how to read a `CloudEvent` starting from a protocol specific data structure.

To convert a `Message` back and forth to `Event`, an `Event` can be wrapped into
a `Message` using [`binding.ToMessage()`](https://github.com/cloudevents/sdk-go/tree/master/v2/binding/event_message.go) and a `Message`
can be converted to `Event` using [`binding.ToEvent()`](https://github.com/cloudevents/sdk-go/tree/master/v2/binding/to_event.go).

A `Message` has its own lifecycle:

* Some implementations of `Message` can be successfully read only one time, 
  because the encoding process drain the message itself. In order to consume a message several 
  times, the [`buffering` module](https://github.com/cloudevents/sdk-go/tree/master/v2/binding/buffering) provides several APIs to buffer the `Message`.
* Every time the `Message` receiver/emitter can forget the message, `Message.Finish()` **must** be invoked.

You can use `Message`s alone or you can interact with them through the protocol implementations.

## `Message` implementation and `Write<DataStructure>` functions

Depending on the protocol binding, the `Message` implementation could support both
binary and structured messages.

All protocol implementations provides a function, with a name like `NewMessage`, to wrap the
[3pl][3pl] data structure into the `Message` implementation. For example, 
`http.NewMessageFromHttpRequest` takes an `net/http.HttpRequest` and wraps it into `protocol/http.Message`, 
which implements `Message`.

`Write<DataStructure>` functions read a `binding.Message` and write what is
found into the [3pl][3pl] data structure. For example, `nats.WriteMsg` writes
the message into a `nats.Msg`, or `http.WriteRequest` writes message into a
`http.Request`.

## Transformations

You can perform simple transformations on `Message` without going through the `Event` representation
using the [`Transformer`s](https://github.com/cloudevents/sdk-go/tree/master/v2/binding/transformer.go).

Some built-in `Transformer`s are provided in the [`transformer` module](https://github.com/cloudevents/sdk-go/tree/master/v2/binding/transformer).

## `protocol` interfaces

Every Protocol implementation provides a set of implemented interfaces to produce/consume messages and 
implement request/response semantics. Six interfaces are defined:

* [`Receiver`](https://github.com/cloudevents/sdk-go/tree/master/v2/protocol/inbound.go): Interface that produces message, receiving them from the wire.
* [`Sender`](https://github.com/cloudevents/sdk-go/tree/master/v2/protocol/outbound.go): Interface that consumes messages, sending them to the wire.
* [`Responder`](https://github.com/cloudevents/sdk-go/tree/master/v2/protocol/inbound.go): Server side request/response.
* [`Requester`](https://github.com/cloudevents/sdk-go/tree/master/v2/protocol/outbound.go): Client side request/response.
* [`Opener`](https://github.com/cloudevents/sdk-go/tree/master/v2/protocol/lifecycle.go): Interface that is optionally needed to bootstrap a `Receiver`/`Responder`
* [`Closer`](https://github.com/cloudevents/sdk-go/tree/master/v2/protocol/lifecycle.go): Interface that is optionally needed to close the connection with remote systems

Every protocol implements one or several of them, depending on its capabilities.

[3pl]: 3pl: 3rd party lib (nats.io or net/http)
