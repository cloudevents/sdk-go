module github.com/cloudevents/sdk-go/protocol/nats_jetstream/v2

go 1.14

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/cloudevents/sdk-go/v2 v2.0.0-00010101000000-000000000000
	github.com/nats-io/nats-server/v2 v2.3.2 // indirect
	github.com/nats-io/nats.go v1.11.1-0.20210623165838-4b75fc59ae30
	google.golang.org/protobuf v1.27.1 // indirect
)
