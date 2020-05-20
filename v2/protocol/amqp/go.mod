module github.com/cloudevents/sdk-go/v2/protocol/amqp

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/Azure/go-amqp v0.12.7
	github.com/cloudevents/sdk-go/v2 v2.0.0-00010101000000-000000000000
	github.com/google/go-cmp v0.4.1 // indirect
	github.com/stretchr/testify v1.5.1
)
