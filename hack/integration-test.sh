#!/usr/bin/env bash

# Copyright 2021 The CloudEvents Authors
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

COVERAGE="`pwd`/coverage.txt"

# ./test/integration only
pushd ./test/integration

# Run integration tests not in parallel
# Prepare coverage file only if not exists
if [ ! -f $COVERAGE ]; then
  touch ./coverage.tmp
  echo 'mode: atomic' > $COVERAGE
fi
COVERPKG=$(go list ./... | grep -v /vendor | tr "\n" ",")
for gomodule in $(go list ./... | grep -v /cmd | grep -v /vendor)
do
  go test -v -parallel 1 -timeout 10m -race -covermode=atomic -coverprofile=coverage.tmp -coverpkg "$COVERPKG" "$gomodule" 2>&1 | sed 's/ of statements in.*//; /warning: no packages being tested depend on matches for pattern /d'
  tail -n +2 coverage.tmp >> $COVERAGE
done
rm coverage.tmp

# Remove test only deps.
go mod tidy

popd
