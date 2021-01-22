module github.com/cloudevents/sdk-go/protocol/nats/v2

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/cloudevents/sdk-go/v2 v2.0.0
	github.com/google/go-cmp v0.4.1 // indirect
	github.com/nats-io/jsm.go v0.0.20
	github.com/nats-io/nats.go v1.10.1-0.20201111151633-9e1f4a0d80d8
)
