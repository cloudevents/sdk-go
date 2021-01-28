module github.com/cloudevents/sdk-go/samples/stan

go 1.14

require (
	github.com/cloudevents/sdk-go/protocol/stan/v2 v2.3.1
	github.com/cloudevents/sdk-go/v2 v2.3.1
)

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/stan/v2 => ../../protocol/stan/v2
