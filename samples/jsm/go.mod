module github.com/cloudevents/sdk-go/samples/jsm

go 1.14

require (
	github.com/cloudevents/sdk-go/protocol/jsm/v2 v2.4.1
	github.com/cloudevents/sdk-go/v2 v2.4.1
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.4.0
)

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/jsm/v2 => ../../protocol/jsm/v2
