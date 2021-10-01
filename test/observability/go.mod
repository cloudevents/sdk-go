module github.com/cloudevents/sdk-go/test/observability

go 1.14

require (
	github.com/cloudevents/sdk-go/observability/opentelemetry/v2 v2.5.0
	github.com/cloudevents/sdk-go/v2 v2.5.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/otel v1.0.0
	go.opentelemetry.io/otel/sdk v1.0.0
	go.opentelemetry.io/otel/trace v1.0.0
)

replace github.com/cloudevents/sdk-go/observability/opentelemetry/v2 => ../../observability/opentelemetry/v2

replace github.com/cloudevents/sdk-go/v2 => ../../v2
