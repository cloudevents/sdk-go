module github.com/cloudevents/sdk-go/protocol/mqtt_paho/v2

go 1.20

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/cloudevents/sdk-go/v2 v2.0.0-00010101000000-000000000000
	github.com/eclipse/paho.golang v0.11.0
)

require (
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
)
