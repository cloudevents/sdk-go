module github.com/cloudevents/sdk-go/samples/ws

go 1.14

require (
	github.com/cloudevents/sdk-go/protocol/ws/v2 v2.4.1
	github.com/cloudevents/sdk-go/v2 v2.4.1
)

replace	github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/ws/v2 => ../../protocol/ws/v2
