#!/usr/bin/env bash

# Copyright 2021 The CloudEvents Authors
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

COVERAGE="`pwd`/coverage.txt"

# test/observability only
pushd ./test/observability

# Prepare coverage file only if not exists
if [ ! -f $COVERAGE ]; then
  touch ./coverage.tmp
  echo 'mode: atomic' > $COVERAGE
fi
COVERPKG="github.com/cloudevents/sdk-go/observability/opentelemetry/v2/..."
for gomodule in $(go list ./... | grep -v /cmd | grep -v /vendor)
do
  go test -v -timeout 30s -race -covermode=atomic -coverprofile=coverage.tmp -coverpkg "$COVERPKG" "$gomodule" 2>&1 | sed 's/ of statements in.*//; /warning: no packages being tested depend on matches for pattern /d'
  tail -n +2 coverage.tmp >> $COVERAGE  
done
rm coverage.tmp

# Remove test only deps.
go mod tidy

popd
