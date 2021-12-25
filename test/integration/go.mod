module github.com/cloudevents/sdk-go/test/integration

go 1.14

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
	github.com/google/go-cmp v0.5.1
	github.com/google/uuid v1.1.1
	github.com/nats-io/nats.go v1.11.1-0.20210623165838-4b75fc59ae30
	github.com/nats-io/stan.go v0.6.0
	github.com/stretchr/testify v1.5.1
	go.uber.org/atomic v1.4.0
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a // indirect
)
