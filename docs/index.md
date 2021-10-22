---
title: Home
nav_order: 1
---

# Golang SDK for CloudEvents

Official CloudEvents SDK to integrate your application with CloudEvents.

This module will help you to:

- Represent CloudEvents in memory
- Use
  [Event Formats](https://github.com/cloudevents/spec/blob/v1.0/spec.md#event-format)
  to serialize/deserialize CloudEvents
- Use
  [Protocol Bindings](https://github.com/cloudevents/spec/blob/v1.0/spec.md#protocol-binding)
  to send/receive CloudEvents

_Note:_ Supported
[CloudEvents specification](https://github.com/cloudevents/spec): 0.3, 1.0

## Get started

Add the module as dependency using go mod:

```
% go get github.com/cloudevents/sdk-go/v2@v2.6.0
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
	if result := c.Send(ctx, event); !cloudevents.IsACK(result) {
		log.Fatalf("failed to send, %v", result)
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
	log.Fatal(c.StartReceiver(context.Background(), receive));
}
```

## Serialize/Deserialize a CloudEvent

To marshal a CloudEvent into JSON:

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

err := json.Unmarshal(bytes, &event)
```

## Supported specification features

|                                                                                         | [v0.3](https://github.com/cloudevents/spec/tree/v0.3) | [v1.0](https://github.com/cloudevents/spec/tree/v1.0) |
| --------------------------------------------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| CloudEvents Core                                                                        | :heavy_check_mark:                                    | :heavy_check_mark:                                    |
| [AMQP Protocol Binding](https://github.com/cloudevents/sdk-go/tree/main/samples/amqp)   | :heavy_check_mark:                                    | :heavy_check_mark:                                    |
| AVRO Event Format                                                                       | :x:                                                   | :x:                                                   |
| [HTTP Protocol Binding](https://github.com/cloudevents/sdk-go/tree/main/samples/http)   | :heavy_check_mark:                                    | :heavy_check_mark:                                    |
| [JSON Event Format](event_data_structure.md#marshalunmarshal-event-to-json)             | :heavy_check_mark:                                    | :heavy_check_mark:                                    |
| [Kafka Protocol Binding](https://github.com/cloudevents/sdk-go/tree/main/samples/kafka) | :heavy_check_mark:                                    | :heavy_check_mark:                                    |
| MQTT Protocol Binding                                                                   | :x:                                                   | :x:                                                   |
| [NATS Protocol Binding](https://github.com/cloudevents/sdk-go/tree/main/samples/nats)   | :heavy_check_mark:                                    | :heavy_check_mark:                                    |
| [STAN Protocol Binding](https://github.com/cloudevents/sdk-go/tree/main/samples/stan)   | :heavy_check_mark:                                    | :heavy_check_mark:                                    |
| Web hook                                                                                | :x:                                                   | :x:                                                   |

## Go further

_. Check out the
[examples](https://github.com/cloudevents/sdk-go/tree/main/samples) _. Dig into
the [Godoc](https://godoc.org/github.com/cloudevents/sdk-go/v2) _. Learn about
the [architecture and concepts](concepts.md) of the SDK _. How to use the
[CloudEvent in-memory representation](event_data_structure.md) \*. How to
use/implement a [Protocol Binding](protocol_implementations.md)
