module github.com/cloudevents/sdk-go/protocol/amqp/v2

go 1.17

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/Azure/go-amqp v0.17.0
	github.com/cloudevents/sdk-go/v2 v2.5.0
	github.com/stretchr/testify v1.8.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.5.1 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
