#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Run integration tests not in parallel
go test -v -parallel 1 ./test/... -coverprofile ${TEST_RESULTS:-.}/integration_test_cover.out -timeout 60s

# Remove test only deps.
go mod tidy
