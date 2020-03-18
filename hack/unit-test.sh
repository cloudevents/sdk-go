#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# v1
pushd ./v1
go test -v ./... -coverprofile ${TEST_RESULTS:-.}/unit_test_cover.out -timeout 15s
# Remove test only deps.
go mod tidy
popd

# v2
pushd ./v2
go test -v ./... -coverprofile ${TEST_RESULTS:-.}/unit_test_cover.out -timeout 15s
# Remove test only deps.
go mod tidy
popd