module github.com/cloudevents/sdk-go/protocol/amqp/v2

go 1.25.0

replace github.com/Azure/go-amqp => github.com/Azure/go-amqp v0.17.0

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/Azure/go-amqp v1.7.0
	github.com/cloudevents/sdk-go/v2 v2.16.2
	github.com/stretchr/testify v1.12.1
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
)
