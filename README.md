# Go SDK for [CloudEvents](https://github.com/cloudevents/spec)

[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/cloudevents/sdk-go/badge)](https://scorecard.dev/viewer/?uri=github.com/cloudevents/sdk-go)
[![go-doc](https://godoc.org/github.com/cloudevents/sdk-go?status.svg)](https://godoc.org/github.com/cloudevents/sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudevents/sdk-go)](https://goreportcard.com/report/github.com/cloudevents/sdk-go)
[![Releases](https://img.shields.io/github/release-pre/cloudevents/sdk-go.svg)](https://github.com/cloudevents/sdk-go/releases)
[![LICENSE](https://img.shields.io/github/license/cloudevents/sdk-go.svg)](https://github.com/cloudevents/sdk-go/blob/main/LICENSE)

Official CloudEvents SDK to integrate your application with CloudEvents.

This library will help you to:

- Represent CloudEvents in memory
- Use
  [Event Formats](https://github.com/cloudevents/spec/blob/v1.0/spec.md#event-format)
  to serialize/deserialize CloudEvents
- Use
  [Protocol Bindings](https://github.com/cloudevents/spec/blob/v1.0/spec.md#protocol-binding)
  to send/receive CloudEvents

_Note:_ Supported
[CloudEvents specification](https://github.com/cloudevents/spec): 0.3, 1.0

_Note:_ Supported go version: 1.22+

## Get started

Add the module as dependency to your project:

```console
go get github.com/cloudevents/sdk-go/v2
```

And import the module in your code

```go
import cloudevents "github.com/cloudevents/sdk-go/v2"
```

## Send your first CloudEvent

To send a CloudEvent using HTTP:

```go
func main() {
	c, err := cloudevents.NewClientHTTP()
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
	if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send, %v", result)
	} else {
		log.Printf("sent: %v", event)
		log.Printf("result: %v", result)
	}
}
```

## Receive your first CloudEvent

To start receiving CloudEvents using HTTP:

```go
func receive(event cloudevents.Event) {
	// do something with event.
    fmt.Printf("%s", event)
}

func main() {
	// The default client is HTTP.
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	if err = c.StartReceiver(context.Background(), receive); err != nil {
		log.Fatalf("failed to start receiver: %v", err)
	}
}
```

## Create a CloudEvent from an HTTP Request

```go
func handler(w http.ResponseWriter, r *http.Request) {
	event, err := cloudevents.NewEventFromHTTPRequest(r)
	if err != nil {
		log.Printf("failed to parse CloudEvent from request: %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
	w.Write([]byte(event.String()))
	log.Println(event.String())
}
```

## Serialize/Deserialize a CloudEvent

To marshal a CloudEvent into JSON:

```go
event := cloudevents.NewEvent()
event.SetID("example-uuid-32943bac6fea")
event.SetSource("example/uri")
event.SetType("example.type")
event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "world"})

bytes, err := json.Marshal(event)
```

To unmarshal JSON back into a CloudEvent:

```go
event :=  cloudevents.NewEvent()

err := json.Unmarshal(bytes, &event)
```

## Go further

- Look at the complete documentation: https://cloudevents.github.io/sdk-go/
- Dig into the godoc: https://godoc.org/github.com/cloudevents/sdk-go/v2
- Check out the [samples directory](./samples) for an extended list of examples
  showing the different SDK features

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
- Email: https://lists.cncf.io/g/cncf-cloudevents-sdk

Each SDK may have its own unique processes, tooling and guidelines, common
governance related material can be found in the
[CloudEvents `community`](https://github.com/cloudevents/spec/tree/master/community)
directory. In particular, in there you will find information concerning how SDK
projects are
[managed](https://github.com/cloudevents/spec/blob/master/community/SDK-GOVERNANCE.md),
[guidelines](https://github.com/cloudevents/spec/blob/master/community/SDK-maintainer-guidelines.md)
for how PR reviews and approval, and our
[Code of Conduct](https://github.com/cloudevents/spec/blob/master/community/GOVERNANCE.md#additional-information)
information.

If there is a security concern with one of the CloudEvents specifications, or
with one of the project's SDKs, please send an email to
[cncf-cloudevents-security@lists.cncf.io](mailto:cncf-cloudevents-security@lists.cncf.io).

## Additional SDK Resources

- [List of current active maintainers](MAINTAINERS.md)
- [How to contribute to the project](CONTRIBUTING.md)
- [SDK's License](LICENSE)
- [SDK's Release process](RELEASING.md)
