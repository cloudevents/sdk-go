module github.com/cloudevents/sdk-go/samples/amqp

go 1.14

require (
	github.com/Azure/go-amqp v0.13.6
	github.com/cloudevents/sdk-go/protocol/amqp/v2 v2.4.0
	github.com/cloudevents/sdk-go/v2 v2.4.0
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.1.1
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/lightstep/tracecontext.go v0.0.0-20181129014701-1757c391b1ac // indirect
	github.com/onsi/ginkgo v1.10.2 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2 // indirect
	golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5 // indirect
)

// Needs a new release.
replace github.com/cloudevents/sdk-go/protocol/amqp/v2 => ../../protocol/amqp/v2