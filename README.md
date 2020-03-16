# Go SDK for [CloudEvents](https://github.com/cloudevents/spec)

[![go-doc](https://godoc.org/github.com/cloudevents/sdk-go?status.svg)](https://godoc.org/github.com/cloudevents/sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudevents/sdk-go)](https://goreportcard.com/report/github.com/cloudevents/sdk-go)
[![CircleCI](https://circleci.com/gh/cloudevents/sdk-go.svg?style=svg)](https://circleci.com/gh/cloudevents/sdk-go)
[![Releases](https://img.shields.io/github/release-pre/cloudevents/sdk-go.svg)](https://github.com/cloudevents/sdk-go/releases)
[![LICENSE](https://img.shields.io/github/license/cloudevents/sdk-go.svg)](https://github.com/cloudevents/sdk-go/blob/master/LICENSE)

## Status

This SDK is still considered work in progress.

**For v1 of the SDK, see** [CloudEvents Go SDK v1](./README_v1.md).

**v2.0.0-preview2:**

In _preview2_ we are focusing on the new Client interface:

```go
type Client interface {
	Send(ctx context.Context, event event.Event) error
	Request(ctx context.Context, event event.Event) (*event.Event, error)
	StartReceiver(ctx context.Context, fn interface{}) error
}
```

Where a full `fn` looks like
`func(context.Context, event.Event) (*event.Event, transport.Result)`

For protocols that do not support responses, `StartReceiver` will throw an error
when attempting to set a receiver fn with that capability.

For protocols that do not support responses from send (Requester interface),
`Client.Request` will throw an error.

**v2.0.0-preview1:**

In _preview1_ we are focusing on the new interfaces found in pkg/transport (will
be renamed to protocol):

- Sender, Send an event.
- Requester, Send an event and expect a response.
- Receiver, Receive an event.
- Responder, Receive an event and respond.

## Working with CloudEvents

_Note:_ Supported
[CloudEvents specification](https://github.com/cloudevents/spec): [0.3, 1.0].

Import this repo to get the `cloudevents` package:

```go
import cloudevents "github.com/cloudevents/sdk-go"
```

To marshal a CloudEvent into JSON, use `event.Event` directly:

```go
event := cloudevents.NewEvent()
event.SetSource("example/uri")
event.SetType("example.type")
event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "world"})

bytes, err := json.Marshal(event)
```

To unmarshal JSON back into a CloudEvent:

```go
event :=  cloudevents.NewEvent()

err := json.Marshal(bytes, &event)
```

The aim of CloudEvents Specification is to define how to "bind" an event to a
particular protocol and back. This SDK wraps the protocol binding
implementations in a client to expose a simple `event.Event` based API.

An example of sending a cloudevents.Event via HTTP:

```go
func main() {
	// The default client is HTTP.
	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	// Create an Event.
	event :=  cloudevents.NewEvent()
	event.SetSource("example/uri")
	event.SetType("example.type")
	event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "world"})

	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	// Send that Event.
	if err := c.Send(ctx, event); err != nil {
		log.Fatalf("failed to send, %v", err)}
	}
}
```

An example of receiving a cloudevents.Event via HTTP:

```go
func receive(event cloudevents.Event) {
	// do something with event.
    fmt.Printf("%s", event)
}

func main() {
	// The default client is HTTP.
	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	log.Fatal(c.StartReceiver(context.Background(), receive));
}
```

Checkout the sample [sender](./cmd/samples/http/sender) and
[receiver](./cmd/samples/http/receiver) applications for working demo.

It can be more performant to not parse an event all the way to the
`event.Event`. For this the package [binding](./pkg/binding) provides primitives
convert `event.Event` to `binding.Message`, and then bind an them onto a
[protocol](./pkg/protocol) implementation.

For example, to convert an `event.Event` to a `binding.Message` and then create
an `http.Request`:

```go
msg := cloudevents.ToMessage(&event)

req, _ = nethttp.NewRequest("POST", "http://localhost", nil)
err = http.WriteRequest(context.TODO(), msg, req, nil)
// ...check error.

// Then use req:
resp, err := http.DefaultClient.Do(req)
```

## Community

- There are bi-weekly calls immediately following the
  [Serverless/CloudEvents call](https://github.com/cloudevents/spec#meeting-time)
  at 9am PT (US Pacific). Which means they will typically start at 10am PT, but
  if the other call ends early then the SDK call will start early as well. See
  the
  [CloudEvents meeting minutes](https://docs.google.com/document/d/1OVF68rpuPK5shIHILK9JOqlZBbfe91RNzQ7u_P7YCDE/edit#)
  to determine which week will have the call.
- Slack: #cloudeventssdk channel under
  [CNCF's Slack workspace](https://slack.cncf.io/).
- Contact for additional information: Scott Nichols (`@Scott Nichols` on slack).
