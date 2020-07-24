module github.com/cloudevents/sdk-go/test/conformance

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/pubsub/v2 => ../../protocol/pubsub/v2

replace github.com/cloudevents/sdk-go/protocol/amqp/v2 => ../../protocol/amqp/v2

replace github.com/cloudevents/sdk-go/protocol/stan/v2 => ../../protocol/stan/v2

replace github.com/cloudevents/sdk-go/protocol/nats/v2 => ../../protocol/nats/v2

replace github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 => ../../protocol/kafka_sarama/v2

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/Azure/go-amqp v0.12.7
	github.com/Shopify/sarama v1.25.0
	github.com/cloudevents/sdk-go v1.2.0
	github.com/cloudevents/sdk-go/protocol/amqp/v2 v2.0.0
	github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 v2.0.0
	github.com/cloudevents/sdk-go/protocol/nats/v2 v2.0.0
	github.com/cloudevents/sdk-go/protocol/pubsub/v2 v2.0.0
	github.com/cloudevents/sdk-go/protocol/stan/v2 v2.0.0
	github.com/cloudevents/sdk-go/v2 v2.0.0
	github.com/cucumber/godog v0.9.0
	github.com/cucumber/messages-go/v10 v10.0.3
	github.com/google/go-cmp v0.4.1
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/stretchr/testify v1.5.1
	go.opencensus.io v0.22.3
)
