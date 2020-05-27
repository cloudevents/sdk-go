#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# test/conformance only
pushd ./test/conformance

go test --tags=conformance -v -timeout 15s

# Remove test only deps.
go mod tidy
popd