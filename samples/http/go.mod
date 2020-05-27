module github.com/cloudevents/sdk-go/samples/http

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../v2

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/cloudevents/sdk-go/v2 v2.0.0
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/kelseyhightower/envconfig v1.4.0
	go.opencensus.io v0.22.3
)
