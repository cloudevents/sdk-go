#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Test everything in pkg and cmd, except amqp
go test -v ./pkg/... ./cmd/... -coverprofile ${TEST_RESULTS:-.}/unit_test_cover.out -timeout 15s

# Remove test only deps.
go mod tidy
