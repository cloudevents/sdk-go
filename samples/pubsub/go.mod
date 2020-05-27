module github.com/cloudevents/sdk-go/samples/pubsub

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/pubsub/v2 => ../../protocol/pubsub/v2

require (
	github.com/cloudevents/sdk-go/protocol/pubsub/v2 v2.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/v2 v2.0.0-00010101000000-000000000000
	github.com/kelseyhightower/envconfig v1.4.0
)
