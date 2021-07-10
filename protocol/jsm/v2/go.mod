module github.com/cloudevents/sdk-go/protocol/jsm/v2

go 1.14

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/cloudevents/sdk-go/v2 v2.0.0-00010101000000-000000000000
	github.com/nats-io/jsm.go v0.0.24
	github.com/nats-io/nats.go v1.11.0
	google.golang.org/protobuf v1.27.1 // indirect
)
