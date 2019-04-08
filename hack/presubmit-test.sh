#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Test everything in pkg and cmd
go test ./pkg/... ./cmd/... -coverprofile cover.out -timeout 15s

# Test everything in test with a slightly longer timeout
go test ./test/... -timeout 60s
