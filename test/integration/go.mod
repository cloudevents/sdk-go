module github.com/cloudevents/sdk-go/test/integration

go 1.23.0

toolchain go1.23.8

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/pubsub/v2 => ../../protocol/pubsub/v2

replace github.com/cloudevents/sdk-go/protocol/amqp/v2 => ../../protocol/amqp/v2

replace github.com/cloudevents/sdk-go/protocol/stan/v2 => ../../protocol/stan/v2

replace github.com/cloudevents/sdk-go/protocol/nats/v2 => ../../protocol/nats/v2

replace github.com/cloudevents/sdk-go/protocol/nats_jetstream/v2 => ../../protocol/nats_jetstream/v2

replace github.com/cloudevents/sdk-go/protocol/nats_jetstream/v3 => ../../protocol/nats_jetstream/v3

replace github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 => ../../protocol/kafka_sarama/v2

replace github.com/cloudevents/sdk-go/protocol/mqtt_paho/v2 => ../../protocol/mqtt_paho/v2

replace github.com/cloudevents/sdk-go/protocol/kafka_confluent/v2 => ../../protocol/kafka_confluent/v2

replace github.com/Azure/go-amqp => github.com/Azure/go-amqp v0.17.0

require (
	github.com/Azure/go-amqp v1.4.0
	github.com/IBM/sarama v1.45.2
	github.com/cloudevents/sdk-go/protocol/amqp/v2 v2.16.1
	github.com/cloudevents/sdk-go/protocol/kafka_confluent/v2 v2.16.0
	github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 v2.16.1
	github.com/cloudevents/sdk-go/protocol/mqtt_paho/v2 v2.16.0
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.16.1
	github.com/cloudevents/sdk-go/protocol/nats_jetstream/v2 v2.16.1
	github.com/cloudevents/sdk-go/protocol/nats_jetstream/v3 v3.0.0-20250721101952-1c64656a6859
	github.com/cloudevents/sdk-go/protocol/stan/v2 v2.16.1
	github.com/cloudevents/sdk-go/v2 v2.16.1
	github.com/confluentinc/confluent-kafka-go/v2 v2.11.0
	github.com/eclipse/paho.golang v0.22.0
	github.com/google/go-cmp v0.7.0
	github.com/google/uuid v1.6.0
	github.com/nats-io/nats.go v1.43.0
	github.com/nats-io/stan.go v0.10.4
	github.com/stretchr/testify v1.10.0
	go.uber.org/atomic v1.11.0
	golang.org/x/sync v0.16.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-hclog v1.1.0 // indirect
	github.com/hashicorp/go-msgpack v1.1.5 // indirect
	github.com/hashicorp/go-msgpack/v2 v2.1.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/hashicorp/raft v1.3.9 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nats-io/jwt/v2 v2.2.1-0.20220113022732-58e87895b296 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
