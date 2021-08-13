module github.com/cloudevents/sdk-go/protocol/ws/v2

go 1.14

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/cloudevents/sdk-go/v2 v2.5.0
	github.com/stretchr/testify v1.5.1
	nhooyr.io/websocket v1.8.6
)
