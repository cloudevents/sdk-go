# Go SDK for [CloudEvents](https://github.com/cloudevents/spec)

[![go-doc](https://godoc.org/github.com/cloudevents/sdk-go?status.svg)](https://godoc.org/github.com/cloudevents/sdk-go)

**NOTE: This SDK is still considered work in progress, things might (and will) 
break with every update.**

## Working with CloudEvents
Package [cloudevents](./pkg/cloudevents) provides primitives to work with 
CloudEvents specification: https://github.com/cloudevents/spec.

Receiving a cloudevents.Event via the Http Transport:

```go
type Receiver struct{}

func (r *Receiver) Receive(event cloudevents.Event) {
	// do something with event.Context and event.Data (via event.DataAs(foo) 
}

func main() {
	t := &cloudeventshttp.Transport{Receiver: &Receiver{}}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", env.Port), t))
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

Sending a cloudevents.Event via the Http Transport with Binary v0.2 encoding:

```go
c, err := client.NewHttpClient(context.TODO(), "http://localhost:8080/", cloudeventshttp.BinaryV02)
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

- [ ] increase `./pkg` code coverage to > 90%. (70% as of Feb 19, 2019)
- [ ] Most tests are happy path, add sad path tests (edge cases).
- [ ] Use contexts to override internal defaults.
- [ ] Fill in Event.Context defaults with values (like ID and time) if 
      nil/empty.
- [x] Might be nice to have the client have a Receive hook.

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

### For v3
- [ ] Support batch json

## Existing Go for CloudEvents

Existing projects that added support for CloudEvents in Go are listed below. 
It's our goal to identify existing patterns of using CloudEvents in Go-based 
project and design the SDK to support these patterns (where it makes sense).
- https://github.com/knative/pkg/tree/master/cloudevents
- https://github.com/vmware/dispatch/blob/master/pkg/events/cloudevent.go
- https://github.com/serverless/event-gateway/tree/master/event
