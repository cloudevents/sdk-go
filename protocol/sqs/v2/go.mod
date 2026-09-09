module github.com/cloudevents/sdk-go/protocol/sqs/v2

go 1.23.0

toolchain go1.23.8

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/aws/aws-sdk-go-v2 v1.38.0
	github.com/aws/aws-sdk-go-v2/service/sqs v1.41.0
	github.com/cloudevents/sdk-go/v2 v2.16.1
)

require (
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.3 // indirect
	github.com/aws/smithy-go v1.22.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
)
