module github.com/cloudevents/sdk-go/samples/amqp

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/amqp/v2 => ../../protocol/amqp/v2

require (
	github.com/Azure/go-amqp v0.12.7
	github.com/cloudevents/sdk-go/protocol/amqp/v2 v2.0.0
	github.com/cloudevents/sdk-go/v2 v2.0.0
	github.com/google/uuid v1.1.1
	go.opencensus.io v0.22.3 // indirect
)
