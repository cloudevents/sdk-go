module github.com/cloudevents/sdk-go/samples/kafka

go 1.23.0

toolchain go1.23.8

require (
	github.com/IBM/sarama v1.45.2
	github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 v2.16.1
	github.com/cloudevents/sdk-go/v2 v2.16.1
	github.com/google/uuid v1.6.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/net v0.41.0 // indirect
)

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 => ../../protocol/kafka_sarama/v2
