module github.com/cloudevents/sdk-go/samples/kafka_confluent

go 1.18

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/kafka_confluent/v2 => ../../protocol/kafka_confluent/v2

require (
	github.com/cloudevents/sdk-go/protocol/kafka_confluent/v2 v2.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/v2 v2.15.2
	github.com/confluentinc/confluent-kafka-go/v2 v2.3.0
)

require (
	github.com/google/uuid v1.3.0 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
)
