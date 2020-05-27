module github.com/cloudevents/sdk-go/protocol/stan/v2

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/cloudevents/sdk-go/v2 v2.0.0
	github.com/google/go-cmp v0.4.1 // indirect
	github.com/nats-io/nats-server/v2 v2.1.7 // indirect
	github.com/nats-io/nats-streaming-server v0.17.0 // indirect
	github.com/nats-io/stan.go v0.6.0
)
