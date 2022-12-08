module github.com/cloudevents/sdk-go/samples/ws

go 1.17

require (
	github.com/cloudevents/sdk-go/protocol/ws/v2 v2.5.0
	github.com/cloudevents/sdk-go/v2 v2.5.0
)

require (
	github.com/google/uuid v1.1.1 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/klauspost/compress v1.15.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
)

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/ws/v2 => ../../protocol/ws/v2
