module github.com/cloudevents/sdk-go/samples/kafka

go 1.14

require (
	github.com/Shopify/sarama v1.25.0
	github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 v2.5.0
	github.com/cloudevents/sdk-go/v2 v2.10.0
	github.com/google/uuid v1.1.1
)

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 => ../../protocol/kafka_sarama/v2
