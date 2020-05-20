module github.com/cloudevents/sdk-go/v2/protocol/kafka_sarama

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/Shopify/sarama v1.19.0
	github.com/cloudevents/sdk-go/v2 v2.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.5.1
)
