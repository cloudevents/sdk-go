# Golang SDK for CloudEvents

To start using the SDK, add the dependency using Go Modules:

```
% go get github.com/cloudevents/sdk-go/v2
```

An example of sending a CloudEvent via HTTP:

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
	if result := c.Send(ctx, event); !cloudevents.IsACK(result) {
		log.Fatalf("failed to send, %v", result)
	}
}
```

An example of receiving a CloudEvent via HTTP:

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

## Supported specification features

|                               |  [v0.3](https://github.com/cloudevents/spec/tree/v0.3) | [v1.0](https://github.com/cloudevents/spec/tree/v1.0) |
| --- | --- | --- |
| CloudEvents Core              | :heavy_check_mark: | :heavy_check_mark: |
| [AMQP Protocol Binding](https://github.com/cloudevents/sdk-go/tree/master/v2/cmd/samples/amqp)         | :heavy_check_mark: | :heavy_check_mark:  |
| AVRO Event Format             | :x: | :x: |
| [HTTP Protocol Binding](https://github.com/cloudevents/sdk-go/tree/master/v2/cmd/samples/http)         | :heavy_check_mark: | :heavy_check_mark: |
| [JSON Event Format](event_data_structure.md##marshalunmarshal-event-to-json)           | :heavy_check_mark: | :heavy_check_mark: |
| [Kafka Protocol Binding](https://github.com/cloudevents/sdk-go/tree/master/v2/cmd/samples/kafka)        | :heavy_check_mark: | :heavy_check_mark: |
| MQTT Protocol Binding         | :x: | :x: |
| [NATS Protocol Binding](https://github.com/cloudevents/sdk-go/tree/master/v2/cmd/samples/nats)         | :heavy_check_mark: | :heavy_check_mark: |
| [STAN Protocol Binding](https://github.com/cloudevents/sdk-go/tree/master/v2/cmd/samples/stan)         | :heavy_check_mark: | :heavy_check_mark: |
| Web hook                      | :x: | :x: |

## Go further

1. [Examples](https://github.com/cloudevents/sdk-go/tree/master/v2/cmd/samples)
1. [Architecture and Concepts](concepts.md)
1. [Event Data Structure documentation](event_data_structure.md)
1. [Protocol implementations](protocol_implementations.md)
