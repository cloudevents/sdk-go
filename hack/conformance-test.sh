#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# v2 only
pushd ./v2/test/conformance/

go test -v -timeout 15s

# Remove test only deps.
go mod tidy
popd