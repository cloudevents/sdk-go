module github.com/cloudevents/sdk-go/sql/v2

go 1.25.0

require (
	github.com/antlr/antlr4/runtime/Go/antlr v1.4.10
	github.com/cloudevents/sdk-go/v2 v2.16.2
	github.com/stretchr/testify v1.12.1
	gopkg.in/yaml.v2 v2.4.0
	sigs.k8s.io/yaml v1.6.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
)

replace github.com/cloudevents/sdk-go/v2 => ../../v2
