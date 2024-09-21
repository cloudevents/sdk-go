module github.com/cloudevents/sdk-go/samples/mqtt

go 1.21

require (
	github.com/cloudevents/sdk-go/protocol/mqtt_paho/v2 v2.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/v2 v2.14.0
	github.com/eclipse/paho.golang v0.21.0
	github.com/google/uuid v1.3.0
)

require (
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
)

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/mqtt_paho/v2 => ../../protocol/mqtt_paho/v2
