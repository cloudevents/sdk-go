module github.com/cloudevents/sdk-go/samples/nats_jetstream

go 1.23.0

toolchain go1.23.8

require (
	github.com/cloudevents/sdk-go/protocol/nats_jetstream/v2 v2.16.0
	github.com/cloudevents/sdk-go/v2 v2.16.0
	github.com/google/uuid v1.6.0
	github.com/kelseyhightower/envconfig v1.4.0
)

require (
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nats-io/nats.go v1.42.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/time v0.0.0-20211116232009-f0f3c7e86c11 // indirect
)

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/nats_jetstream/v2 => ./../../protocol/nats_jetstream/v2
