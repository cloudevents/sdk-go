# Go SDK for [CloudEvents](https://github.com/cloudevents/spec)

**NOTE: This SDK is still considered work in progress, things might (and will) break with every update.**

Package cloudevents provides primitives to work with CloudEvents specification: https://github.com/cloudevents/spec.

Parsing Event from HTTP Request:
```go
import "github.com/cloudevents/sdk-go"
	marshaller := v01.NewDefaultHTTPMarshaller()
	// req is *http.Request
	event, err := marshaller.FromRequest(req)
	if err != nil {
		panic("Unable to parse event from http Request: " + err.String())
	}
	fmt.Printf("eventType: %s", event.Get("eventType")
```

Creating a minimal CloudEvent in version 0.1:
```go
import "github.com/cloudevents/sdk-go/v01"
	event := v01.Event{
		EventType:        "com.example.file.created",
		Source:           "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
		EventID:          "ea35b24ede421",
	}
```

Creating HTTP request from CloudEvent:
```
marshaller := v01.NewDefaultHTTPMarshaller()
var req *http.Request
err := event.ToRequest(req)
if err != nil {
	panic("Unable to marshal event into http Request: " + err.String())
}
```

The goal of this package is to provide support for all released versions of CloudEvents, ideally while maintaining
the same API. It will use semantic versioning with following rules:
* MAJOR version increments when backwards incompatible changes is introduced.
* MINOR version increments when backwards compatible feature is introduced INCLUDING support for new CloudEvents version.
* PATCH version increments when a backwards compatible bug fix is introduced.


## TODO list

- [ ] Add encoders registry, where SDK user can register their custom content-type encoders/decoders
- [ ] Add more tests for edge cases

## Existing Go for CloudEvents

Existing projects that added support for CloudEvents in Go are listed below. It's our goal to identify existing patterns
of using CloudEvents in Go-based project and design the SDK to support these patterns (where it makes sense).
- https://github.com/knative/eventing/tree/master/pkg/event
- https://github.com/vmware/dispatch/blob/master/pkg/events/cloudevent.go
- https://github.com/serverless/event-gateway/tree/master/event
