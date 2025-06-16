module github.com/cloudevents/sdk-go/samples/kafka_confluent

go 1.23.0

toolchain go1.23.8

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/kafka_confluent/v2 => ./../../protocol/kafka_confluent/v2

require (
	github.com/cloudevents/sdk-go/protocol/kafka_confluent/v2 v2.16.0
	github.com/cloudevents/sdk-go/v2 v2.16.1
	github.com/confluentinc/confluent-kafka-go/v2 v2.10.1
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
)
