# Development

This is a collection of topics related to development of the SDK.

## Transcoding

One of the goals of this sdk is to decouple the encoding from the transport as
much as possible. This is done by the sdk integrator interacting primarily with
the [Event][cloudevents.event] object. The [Transport][transport.transport]
implementation interacts with a [Message][transport.message] object. The
[Codec][transport.codec] is responsible for converting `Event` to `Message` for
the the transport implementation. And this process works in reverse when a
transport is used to receive an event.

Sending:

```
 Event -[via Codec]-> Message -> Transport
```

Receiving:

```
(Transport) -> Message -[via Codec]-> Event
```

The CloudEvents spec outlines the various ways encoding is allowed, and there
are two levels of encoding.

1. Encoding of the context and extension attributes of the event.
1. Encoding of the data payload of the event.

For this reason there is also a [DataCodec][datacodec.codec] that is responsible
for converting an encoded data payload into the intended format. These formats
tend to be `application/xml`, `text/xml`, `application/json`, and `text/json`.

The DataCodec is invoked (as of this writing) when `event.SetData` is called, or
when `event.DataAs` is called when receiving.

[cloudevents.event]: ../pkg/cloudevents/event.go
[transport.transport]: ../pkg/cloudevents/transport/transport.go
[transport.message]: ../pkg/cloudevents/transport/message.go
[transport.codec]: ../pkg/cloudevents/transport/codec.go
[datacodec.codec]: ../pkg/cloudevents/datacodec/codec.go
