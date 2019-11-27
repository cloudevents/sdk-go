# Go SDK for [CloudEvents](https://github.com/cloudevents/spec)

[![go-doc](https://godoc.org/github.com/cloudevents/sdk-go?status.svg)](https://godoc.org/github.com/cloudevents/sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudevents/sdk-go)](https://goreportcard.com/report/github.com/cloudevents/sdk-go)
[![CircleCI](https://circleci.com/gh/cloudevents/sdk-go.svg?style=svg)](https://circleci.com/gh/cloudevents/sdk-go)
[![Releases](https://img.shields.io/github/release-pre/cloudevents/sdk-go.svg)](https://github.com/cloudevents/sdk-go/releases)
[![LICENSE](https://img.shields.io/github/license/cloudevents/sdk-go.svg)](https://github.com/cloudevents/sdk-go/blob/master/LICENSE)

**NOTE: This SDK is still considered work in progress, things might (and will)
break with every update.**

## Working with CloudEvents

Package [cloudevents](./pkg/cloudevents) provides primitives to work with
CloudEvents specification: https://github.com/cloudevents/spec.

Import this repo to get the `cloudevents` package:

```go
import "github.com/cloudevents/sdk-go"
```

Receiving a cloudevents.Event via the HTTP Transport:

```go
func Receive(event cloudevents.Event) {
	baz
}

func main() {
	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	log.Fatal(c.StartReceiver(context.Background(), Receive));
}
```

Creating a minimal CloudEvent in version 0.2:

```go
event := cloudevents.NewEvent()
event.SetID("ABC-123")
event.SetType("com.cloudevents.readme.sent")
event.SetSource("http://localhost:8080/")
event.SetData(data)
```

Sending a cloudevents.Event via the HTTP Transport with Binary v0.2 encoding:

```go
t, err := cloudevents.NewHTTPTransport(
	cloudevents.WithTarget("http://localhost:8080/"),
	cloudevents.WithEncoding(cloudevents.HTTPBinaryV02),
)
if err != nil {
	panic("failed to create transport, " + err.Error())
}

c, err := cloudevents.NewClient(t)
if err != nil {
	panic("unable to create cloudevent client: " + err.Error())
}
if err := c.Send(ctx, event); err != nil {
	panic("failed to send cloudevent: " + err.Error())
}
```

Or, the transport can be set to produce CloudEvents using the selected encoding
but not change the provided event version, here the client is set to output
structured encoding:

```go
t, err := cloudevents.NewHTTPTransport(
	cloudevents.WithTarget("http://localhost:8080/"),
	cloudevents.WithStructuredEncoding(),
)
```

If you are using advanced transport features or have implemented your own
transport integration, provide it to a client so your integration does not
change:

```go
t, err := cloudevents.NewHTTPTransport(
	cloudevents.WithPort(8181),
	cloudevents.WithPath("/events/")
)
// or a custom transport: t := &custom.MyTransport{Cool:opts}

c, err := cloudevents.NewClient(t, opts...)
```

Checkout the sample [sender](./cmd/samples/http/sender) and
[receiver](./cmd/samples/http/receiver) applications for working demo.

## Client Options

There are several client options that can be passed in when making a CloudEvents
client. These help with defaults or validation that could be useful for a
particular need. There is also hooks to add your own custom options and defaults
to extend the client provided by the Golang CloudEvents SDK.

### WithEventDefaulter

`WithEventDefaulter` is the generic hook to add a defaulter to the defaulter
chain. With the following function type:

```go
type EventDefaulter func(ctx context.Context, event cloudevents.Event) cloudevents.Event
```

Implement a `EventDefaulter` function:

```go
func customOption(ctx context.Context, event cloudevents.Event) cloudevents.Event {
	// TODO(reader): mutate event.
    return event
}
```

And then pass it into the client:

```go
cloudevents.NewClient(t, customOption)
```

### WithUUIDs

`WithUUIDs` sets `event.Context.ID` to a new UUID if `ID` is not set.

### WithTimeNow

`WithTimeNow` sets `event.Context.Time` to a `time.Now()` if `Time` is not set.

### WithConverterFn

`WithConverterFn` allows you to introduce a function to give one last try to
convert a non-CloudEvent into a CloudEvent for supported transports. The convert
function signature should be:

```go
func (ctx context.Context, m transport.Message, err error) (*cloudevents.Event, error)
```

See the [converter](./cmd/samples/http/converter/receiver) sample for a working
example.

### WithOverrides

`WithOverrides` allows you to create a set of files on the filesystem that will
be watched, read, and mutate the outbound event.

File names will be used as the extension attribute name (only extensions are
allowed, no first-class attributes name). File contents will be used as the
value of the extension attribute.

For example:

```bash
$ tree cmd/samples/http/overrides/extensions/
cmd/samples/http/overrides/extensions/
├── baz
└── foo

$ cat cmd/samples/http/overrides/extensions/baz
bar

$ cat cmd/samples/http/overrides/extensions/foo
42
```

`baz` and `foo` will be added to the CloudEvents extensions with "bar" and "42",
respectively.

See the [overrides](./cmd/samples/http/overrides) sample for a working example.
