module github.com/cloudevents/sdk-go/samples/nats

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/nats/v2 => ../../protocol/nats

require (
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/v2 v2.0.0-00010101000000-000000000000
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.4.0
)
