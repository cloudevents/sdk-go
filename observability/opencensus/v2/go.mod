module github.com/cloudevents/sdk-go/observability/opencensus/v2

go 1.17

require (
	github.com/cloudevents/sdk-go/v2 v2.5.0
	github.com/lightstep/tracecontext.go v0.0.0-20181129014701-1757c391b1ac
	github.com/stretchr/testify v1.8.0
	go.opencensus.io v0.22.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	github.com/onsi/ginkgo v1.14.2 // indirect
	github.com/onsi/gomega v1.10.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/cloudevents/sdk-go/v2 => ../../../v2
