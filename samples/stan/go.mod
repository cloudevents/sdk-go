module github.com/cloudevents/sdk-go/samples/stan

go 1.13

replace github.com/cloudevents/sdk-go/protocol/stan/v2 => ../../protocol/stan/v2

require (
	github.com/cloudevents/sdk-go/protocol/stan/v2 v2.0.0
	github.com/cloudevents/sdk-go/v2 v2.1.0
)
