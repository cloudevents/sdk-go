module github.com/cloudevents/sdk-go/test/benchmark

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/protocol/pubsub/v2 => ../../protocol/pubsub/v2

replace github.com/cloudevents/sdk-go/protocol/amqp/v2 => ../../protocol/amqp/v2

replace github.com/cloudevents/sdk-go/protocol/stan/v2 => ../../protocol/stan/v2

replace github.com/cloudevents/sdk-go/protocol/nats/v2 => ../../protocol/nats/v2

replace github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 => ../../protocol/kafka_sarama/v2

require (
	github.com/Shopify/sarama v1.25.0
	github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2 v2.1.0
	github.com/cloudevents/sdk-go/v2 v2.1.0
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/go-cmp v0.4.1 // indirect
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/crypto v0.0.0-20200323165209-0ec3e9974c59 // indirect
	golang.org/x/exp v0.0.0-20200224162631-6cc2880d07d6 // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/sys v0.0.0-20200331124033-c3d80250170d // indirect
	gonum.org/v1/netlib v0.0.0-20200603212716-16abd5ac5bc7 // indirect
	gonum.org/v1/plot v0.7.0
	google.golang.org/protobuf v1.22.0 // indirect
)
