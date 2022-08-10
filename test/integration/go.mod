module github.com/cloudevents/sdk-go/test/integration

go 1.17

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/pubsub/v2 => ../../protocol/pubsub/v2

replace github.com/cloudevents/sdk-go/protocol/amqp/v2 => ../../protocol/amqp/v2

replace github.com/cloudevents/sdk-go/protocol/stan/v2 => ../../protocol/stan/v2

replace github.com/cloudevents/sdk-go/protocol/nats/v2 => ../../protocol/nats/v2

replace github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 => ../../protocol/kafka_sarama/v2

require (
	github.com/Azure/go-amqp v0.17.0
	github.com/Shopify/sarama v1.25.0
	github.com/cloudevents/sdk-go/protocol/amqp/v2 v2.5.0
	github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 v2.5.0
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.5.0
	github.com/cloudevents/sdk-go/protocol/stan/v2 v2.5.0
	github.com/cloudevents/sdk-go/v2 v2.5.0
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.1.1
	github.com/nats-io/nats.go v1.11.1-0.20210623165838-4b75fc59ae30
	github.com/nats-io/stan.go v0.6.0
	github.com/stretchr/testify v1.8.0
	go.uber.org/atomic v1.4.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/hashicorp/go-msgpack v1.1.5 // indirect
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/hashicorp/raft v1.3.9 // indirect
	github.com/jcmturner/gofork v0.0.0-20190328161633-dc7c13fece03 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/klauspost/compress v1.11.12 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	github.com/nats-io/jwt/v2 v2.2.0 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/jcmturner/aescts.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/dnsutils.v1 v1.0.1 // indirect
	gopkg.in/jcmturner/gokrb5.v7 v7.2.3 // indirect
	gopkg.in/jcmturner/rpc.v1 v1.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
