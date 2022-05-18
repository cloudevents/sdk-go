module github.com/cloudevents/sdk-go/samples/ws

go 1.14

require (
	github.com/cloudevents/sdk-go/protocol/ws/v2 v2.5.0
	github.com/cloudevents/sdk-go/v2 v2.10.0
)

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/ws/v2 => ../../protocol/ws/v2
