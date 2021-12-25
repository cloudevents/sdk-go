module github.com/cloudevents/sdk-go/samples/amqp

go 1.14

require (
	github.com/Azure/go-amqp v0.17.0
	github.com/cloudevents/sdk-go/protocol/amqp/v2 v2.5.0
	github.com/cloudevents/sdk-go/v2 v2.5.0
	github.com/google/uuid v1.1.1
)

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/amqp/v2 => ../../protocol/amqp/v2
