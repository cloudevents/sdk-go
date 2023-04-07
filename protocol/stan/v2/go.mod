module github.com/cloudevents/sdk-go/protocol/stan/v2

go 1.17

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/cloudevents/sdk-go/v2 v2.5.0
	github.com/nats-io/stan.go v0.10.2
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	github.com/nats-io/nats-server/v2 v2.7.4 // indirect
	github.com/nats-io/nats-streaming-server v0.24.3 // indirect
	github.com/nats-io/nats.go v1.13.1-0.20220308171302-2f2f6968e98d // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20220307211146-efcb8507fb70 // indirect
)
