module github.com/cloudevents/sdk-go/samples/ws

go 1.14

replace (
	github.com/cloudevents/sdk-go/protocol/ws/v2 => ../../protocol/ws/v2
	github.com/cloudevents/sdk-go/v2 => ../../v2
)

require (
	github.com/cloudevents/sdk-go/protocol/ws/v2 v2.0.0
	github.com/cloudevents/sdk-go/v2 v2.0.0
)
