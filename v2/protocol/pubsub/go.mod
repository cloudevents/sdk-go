module github.com/cloudevents/sdk-go/v2/protocol/pubsub

go 1.13

replace github.com/cloudevents/sdk-go/v2 => ../../../v2

require (
	cloud.google.com/go/pubsub v1.3.1
	github.com/cloudevents/sdk-go/v2 v2.0.0-00010101000000-000000000000
	github.com/google/go-cmp v0.4.1
	google.golang.org/api v0.24.0
	google.golang.org/grpc v1.29.1
)
