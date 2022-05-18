module github.com/cloudevents/sdk-go/samples/nats

go 1.14

require (
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.5.0
	github.com/cloudevents/sdk-go/v2 v2.10.0
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.4.0
)

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/nats/v2 => ../../protocol/nats/v2
