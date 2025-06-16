module github.com/cloudevents/sdk-go/samples/stan

go 1.23.0

toolchain go1.23.8

require (
	github.com/cloudevents/sdk-go/protocol/stan/v2 v2.16.1
	github.com/cloudevents/sdk-go/v2 v2.16.1
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-hclog v1.1.0 // indirect
	github.com/hashicorp/go-msgpack v1.1.5 // indirect
	github.com/hashicorp/go-msgpack/v2 v2.1.3 // indirect
	github.com/hashicorp/raft v1.3.9 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nats-io/jwt/v2 v2.2.1-0.20220113022732-58e87895b296 // indirect
	github.com/nats-io/nats.go v1.43.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nats-io/stan.go v0.10.4 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/stan/v2 => ../../protocol/stan/v2
