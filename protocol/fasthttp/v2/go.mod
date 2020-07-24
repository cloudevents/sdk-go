module github.com/cloudevents/sdk-go/protocol/fasthttp/v2

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	github.com/cloudevents/sdk-go/v2 v2.1.0
	github.com/google/go-cmp v0.5.1
	github.com/stretchr/testify v1.6.1
	github.com/valyala/fasthttp v1.15.1
	go.opencensus.io v0.22.4
	go.uber.org/zap v1.15.0
)
