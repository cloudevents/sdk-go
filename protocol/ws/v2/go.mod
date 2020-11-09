module github.com/cloudevents/sdk-go/protocol/ws/v2

go 1.14

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/cloudevents/sdk-go/v2 v2.0.0
	nhooyr.io/websocket v1.8.6 // indirect
)
