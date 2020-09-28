module github.com/cloudevents/sdk-go/protocol/amqp/v2

go 1.14

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/Azure/go-amqp v0.12.7
	github.com/cloudevents/sdk-go/v2 v2.0.0
	github.com/google/go-cmp v0.4.1 // indirect
	github.com/stretchr/testify v1.5.1
)
