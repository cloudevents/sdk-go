# Development

This is a collection of topics related to development of the SDK.

## Datacodec

The CloudEvents spec has different encoding rules for event attributes and event
data.

Package [cloudevents/types][cloudevents.types] implements the CE type system for
attribute values.

Package [datacodec/codec][datacodec.codec] is responsible for encoding and
decoding the data payload. Some formats have built-in support e.g.
`application/xml` and `application/json`.

The DataCodec is invoked (as of this writing) when `event.SetData` is called, or
when `event.DataAs` is called when receiving.

event.Context.SetDataContenType() can also be set to a content-type that is not
known to the datacodec. In that case Event.Data should be set or read directly
as encoded []byte or string data, which will not be interpreted by this library.

## Transcoding and Forwarding

One of the goals of this sdk is to decouple the encoding from the protocol
binding as much as possible, and to enable forwarding of messages from one
protocol to another without loss of efficiency or reliability.

Protocol bindings implement the binding.Message API to achieve this. The event
can be represented in two ways: in "structured mode" as a (mediaType string,
[]byte) pair, or decoded as a "binary mode" [Event][cloudevents.event] object.

Structured mode allows forwarding structured events with minimal re-encoding,
but the event is a "black-box", it's attributes and data are not accessible.
Binary mode requires decoding and re-encoding of protocol messages, but allows
the event to be examined and modified by the application.

In-memory messages in both modes can be created using the binding package:

Sending:

```
 var e cloudevents.Event

 // Binary message:
 sender.Send(ctx, binding.EventMessage(e))

 // Pre-encoded structured message:
 sender.Send(ctx, binding.StructMessage{Format: "media-type", Bytes: bytes})

 // Format a binary Event as a structued message:
 var f format.Format = ...
 bytes, err := f.Marshal(e)
 m, err := binding.StructuredEncoder{Format: f.MediaType(), Bytes: bytes}
 sender.Send(ctx, m)
```

Receiving:

```
m, err := receiver.Receive(ctx)

// Extract structured event if it is present
if format, bytes := m.Structured(); format != nil { /* use structured message */ }

// Decode as a binary Event
e, err := m.Event()
```

The interface binding.Message also provides generic methods to handle reliable
delivery Qualities of Service between different protocols, provided both
protocols support the required QoS.

See pkg/binding/doc.go for more details.

[cloudevents.event]: ../pkg/cloudevents/event.go
[cloudevents.types]: ../pkg/types/doc.go
[transport.transport]: ../pkg/cloudevents/transport/transport.go
[transport.message]: ../pkg/cloudevents/transport/message.go
[transport.codec]: ../pkg/cloudevents/transport/codec.go
[datacodec.codec]: ../pkg/event/datacodec/codec.go
