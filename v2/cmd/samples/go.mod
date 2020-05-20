module github.com/cloudevents/sdk-go/v2/cmd/samples

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

replace github.com/cloudevents/sdk-go/v2/protocol/pubsub => ../../../v2/protocol/pubsub

replace github.com/cloudevents/sdk-go/v2/protocol/amqp => ../../../v2/protocol/amqp

replace github.com/cloudevents/sdk-go/v2/protocol/stan => ../../../v2/protocol/stan

replace github.com/cloudevents/sdk-go/v2/protocol/nats => ../../../v2/protocol/nats

replace github.com/cloudevents/sdk-go/v2/protocol/kafka_sarama => ../../../v2/protocol/kafka_sarama

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/Azure/go-amqp v0.12.7
	github.com/Shopify/sarama v1.19.0
	github.com/Shopify/toxiproxy v2.1.4+incompatible // indirect
	github.com/cloudevents/sdk-go/v2 v2.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/v2/protocol/amqp v0.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/v2/protocol/kafka_sarama v0.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/v2/protocol/nats v0.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/v2/protocol/pubsub v0.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/v2/protocol/stan v0.0.0-00010101000000-000000000000
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/frankban/quicktest v1.10.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/nats-io/nats-server/v2 v2.1.7 // indirect
	github.com/nats-io/nats-streaming-server v0.17.0 // indirect
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0 // indirect
	go.opencensus.io v0.22.3
)
