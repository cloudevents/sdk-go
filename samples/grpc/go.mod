module github.com/cloudevents/sdk-go/samples/grpc

go 1.18

require (
	github.com/cloudevents/sdk-go/binding/format/protobuf/v2 v2.14.0
	github.com/cloudevents/sdk-go/protocol/grpc v0.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/v2 v2.14.0
	github.com/google/uuid v1.3.1
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230822172742-b8732ec3820d // indirect
)

replace github.com/cloudevents/sdk-go/v2 => ../../v2

replace github.com/cloudevents/sdk-go/binding/format/protobuf/v2 => ../../binding/format/protobuf/v2

replace github.com/cloudevents/sdk-go/protocol/grpc => ../../protocol/grpc
