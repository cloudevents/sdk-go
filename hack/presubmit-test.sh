#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Test everything in pkg and cmd, except amqp
go test -v ./pkg/... ./cmd/... -coverprofile cover.out -timeout 15s

# AMQP & Kafka cannot run tests in parallel
go test -v -parallel 1 -tags amqp ./pkg/bindings/amqp -coverprofile amqp_bindings_cover.out -timeout 15s
go test -v -parallel 1 -tags amqp ./pkg/cloudevents/transport/amqp -coverprofile amqp_transport_cover.out -timeout 15s
go test -v -parallel 1 -tags kafka ./pkg/bindings/amqp -coverprofile kafka_bindings_cover.out -timeout 15s

# Test everything in test with a slightly longer timeout
go test ./test/... -timeout 60s

# Remove test only deps.
go mod tidy
