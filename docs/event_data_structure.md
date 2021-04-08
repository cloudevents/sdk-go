---
title: Event data structure
nav_order: 3
---

# Event data structure

{: .no_toc }

1. TOC {:toc}

## Overview

The `Event` data structure is the representation of a `CloudEvent` in sdk-go.
The main features are:

- Support for multiple event versions
- Type safe implementation of CloudEvents attributes and extensions
- Validation of the event
- Implementation of marshalling/unmarshalling to JSON.
- Implementation of `data` field codecs

## Writing/Reading the attributes and extensions

To read and write attributes and extensions of the `Event`, you can use the
methods of `EventContextReader`/`EventContextWriter`:

```go
ev := cloudevents.NewEvent()
err := ev.Context.SetSource("http://localhost")
```

```go
ev := cloudevents.NewEvent()
err := ev.Context.SetExtension("aaa", "hello_world")
```

Attributes and extensions are represented internally using wrapper types from
[`types` module](https://github.com/cloudevents/sdk-go/tree/main/v2/types).

## Writing/Reading the data field

To write the `data` field in your `Event`, use `Event.SetData()`. This method
accepts both the content type and the payload. If the payload is a `[]byte`,
then no encoding will be made, otherwise the
[`datacodec` module](https://github.com/cloudevents/sdk-go/tree/main/v2/event/datacodec)
will be used to encode the payload.

You can read the `data` or accessing directly to the underlying `[]byte` using
`Event.Data()` or decoding it using `event.DataAs()`, which uses a specific
`Decoder` from `datacodec` module to decode the event `data`.

Some formats have built-in support in `datacodec`, like `application/xml`,
`application/json` and `text/plain`. You can use your own encoding implementing
[`datacodec.Encoder` and `datacodec.Decoder`](https://github.com/cloudevents/sdk-go/tree/main/v2/event/datacodec/codec.go)
and registering it with
[`datacodec.AddEncoder` and `datacodec.AddDecoder`](https://github.com/cloudevents/sdk-go/tree/main/v2/event/datacodec/codec.go)

## Marshal/Unmarshal event to JSON

To marshal the Event to JSON:

```go
ev := cloudevents.NewEvent()
bytesArray, err := json.Marshal(event)
```

To unmarshal the Event from JSON:

```go
ev := &event.Event{}
err := json.Unmarshal(bytesArray, ev)
```
