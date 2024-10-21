module github.com/cloudevents/sdk-go/observability/opencensus/v2

go 1.22

require (
	github.com/cloudevents/sdk-go/v2 v2.5.0
	github.com/lightstep/tracecontext.go v0.0.0-20181129014701-1757c391b1ac
	github.com/stretchr/testify v1.9.0
	go.opencensus.io v0.22.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/onsi/ginkgo v1.14.2 // indirect
	github.com/onsi/gomega v1.10.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/cloudevents/sdk-go/v2 => ../../../v2
