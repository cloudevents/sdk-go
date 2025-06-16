module github.com/cloudevents/sdk-go/protocol/stan/v2

go 1.23.0

toolchain go1.23.8

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/cloudevents/sdk-go/v2 v2.16.1
	github.com/nats-io/stan.go v0.10.4
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nats-io/nats-server/v2 v2.10.27 // indirect
	github.com/nats-io/nats-streaming-server v0.25.6 // indirect
	github.com/nats-io/nats.go v1.43.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)
