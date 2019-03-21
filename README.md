# Go SDK for [CloudEvents](https://github.com/cloudevents/spec)

[![go-doc](https://godoc.org/github.com/cloudevents/sdk-go?status.svg)](https://godoc.org/github.com/cloudevents/sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudevents/sdk-go)](https://goreportcard.com/report/github.com/cloudevents/sdk-go)
[![CircleCI](https://circleci.com/gh/cloudevents/sdk-go.svg?style=svg)](https://circleci.com/gh/cloudevents/sdk-go)

**NOTE: This SDK is still considered work in progress, things might (and will)
break with every update.**

## Working with CloudEvents

Package [cloudevents](./pkg/cloudevents) provides primitives to work with
CloudEvents specification: https://github.com/cloudevents/spec.

Receiving a cloudevents.Event via the HTTP Transport:

```go
func Receive(event cloudevents.Event) {
	// do something with event.Context and event.Data (via event.DataAs(foo)
}

func main() {
	c, err := client.NewDefault()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	log.Fatal(c.StartReceiver(context.Background(), Receive));
}
```

Creating a minimal CloudEvent in version 0.2:

```go
event := cloudevents.Event{
	Context: cloudevents.EventContextV02{
		ID:     uuid.New().String(),
		Type:   "com.cloudevents.readme.sent",
		Source: types.ParseURLRef("http://localhost:8080/"),
	}.AsV02(),
}
```

Sending a cloudevents.Event via the HTTP Transport with Binary v0.2 encoding:

```go
// import cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"

t, err := cloudeventshttp.New(
	cehttp.WithTarget("http://localhost:8080/"),
	cehttp.WithEncoding(cehttp.BinaryV02),
)
if err != nil {
	panic("failed to create transport, " + err.Error())
}

c, err := client.New(t)
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
// import cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"

t, err := cloudeventshttp.New(
	cehttp.WithTarget("http://localhost:8080/"),
	cehttp.WithStructuredEncoding(),
)
```

If you are using advanced transport features or have implemented your own
transport integration, provide it to a client so your integration does not
change:

```go
// import (
//   "github.com/cloudevents/sdk-go/pkg/cloudevents/client"
//   cehttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
// )

t, err := transporthttp.New(cehttp.WithPort(8181), cehttp.WithPath("/events/"))
// or a custom transport: t := &custom.MyTransport{Cool:opts}

c, err := client.New(t, opts...)
```

Checkout the sample [sender](./cmd/samples/http/sender) and
[receiver](./cmd/samples/http/receiver) applications for working demo.
