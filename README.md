# Go SDK for [CloudEvents](https://github.com/cloudevents/spec)

[![go-doc](https://godoc.org/github.com/cloudevents/sdk-go?status.svg)](https://godoc.org/github.com/cloudevents/sdk-go)

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
	_, err := client.StartHttpReceiver(Receive)
	if err != nil {
		log.Fatal(err)
	}
	<-context.Background().Done()
}
```

Creating a minimal CloudEvent in version 0.2:

```go
event := cloudevents.Event{
    Context: cloudevents.EventContextV02{
        ID:     uuid.New().String(),
        Type:   "com.cloudevents.readme.sent",
        Source: types.ParseURLRef("http://localhost:8080/"),
    },
}
```

Sending a cloudevents.Event via the HTTP Transport with Binary v0.2 encoding:

```go
c, err := client.NewHttpClient(
	client.WithTarget("http://localhost:8080/"),
	client.WithHttpEncoding(cloudeventshttp.BinaryV02), 
)
if err != nil {
	panic("unable to create cloudevent client: " + err.Error())
}
if err := c.Send(event); err != nil {
	panic("failed to send cloudevent: " + err.Error())
}
```

Checkout the sample [sender](./cmd/samples/sender) and 
[receiver](./cmd/samples/receiver) applications for working demo. 

## TODO list

### General 

- [ ] Add details to what the samples are showing.
- [ ] Add a sample to show how to use the transport without the client.
- [ ] increase `./pkg` code coverage to > 90%. (70% as of Feb 19, 2019)
- [ ] Most tests are happy path, add sad path tests (edge cases).
- [ ] Use contexts to override internal defaults.
- [ ] Fill in Event.Context defaults with values (like ID and time) if 
      nil/empty.
- [x] Might be nice to have the client have a Receive hook.
- [ ] Might be an issue with zero body length requests.
- [ ] Need a change to the client to make making events easier
- [ ] Implement String() for event context

### Webhook
- [ ] Implement Auth in webhook
- [ ] Implement Callback in webhook
- [ ] Implement Allowed Origin
- [ ] Implement Allowed Rate

### JSON
- [ ] Support json value as body. 

### HTTP
- [ ] Support overrides for method.
- [ ] Merge headers from context on send.

### Nats
- [ ] Plumb in auth for the nats server.
- [ ] v0.2 and v0.3 are very similar. Combine decode logic?

### For v0.3
- [ ] Support batch json

## Existing Go for CloudEvents

Existing projects that added support for CloudEvents in Go are listed below. 
It's our goal to identify existing patterns of using CloudEvents in Go-based 
project and design the SDK to support these patterns (where it makes sense).
- https://github.com/serverless/event-gateway/tree/master/event
