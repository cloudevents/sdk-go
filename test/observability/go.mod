module github.com/cloudevents/sdk-go/test/observability

go 1.22

require (
	github.com/cloudevents/sdk-go/observability/opentelemetry/v2 v2.5.0
	github.com/cloudevents/sdk-go/v2 v2.5.0
	github.com/stretchr/testify v1.10.0
	go.opentelemetry.io/otel v1.18.0
	go.opentelemetry.io/otel/sdk v1.0.0
	go.opentelemetry.io/otel/trace v1.18.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.18.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/sys v0.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/cloudevents/sdk-go/observability/opentelemetry/v2 => ../../observability/opentelemetry/v2

replace github.com/cloudevents/sdk-go/v2 => ../../v2
