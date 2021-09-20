module github.com/cloudevents/sdk-go/observability/opentelemetry/v2

go 1.14

require (
	github.com/cloudevents/sdk-go/v2 v2.5.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.23.0
	go.opentelemetry.io/otel v1.0.0-RC3
	go.opentelemetry.io/otel/trace v1.0.0-RC3
)

replace github.com/cloudevents/sdk-go/v2 => ../../../v2
