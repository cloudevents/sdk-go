module github.com/cloudevents/sdk-go/protocol/rocketmq/v2

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/apache/rocketmq-client-go/v2 v2.1.0-rc3
	github.com/cloudevents/sdk-go/v2 v2.0.0
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/frankban/quicktest v1.10.0 // indirect
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0 // indirect
	github.com/stretchr/testify v1.5.1
)
