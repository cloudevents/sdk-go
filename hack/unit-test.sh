#!/usr/bin/env bash

# Copyright 2021 The CloudEvents Authors
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

COVERAGE="`pwd`/coverage.txt"
echo 'mode: atomic' > $COVERAGE

for gomodule in $(find . | grep "go\.mod" | awk '{gsub(/\/go.mod/,""); print $0}' | grep -v "./test" | grep -v "./conformance")
do
  echo
  echo --- Testing $gomodule ---
  echo
  
  pushd $gomodule
  touch ./coverage.tmp
  COVERPKG=$(go list ./... | grep -v /vendor | grep -v /test | tr "\n" ",")

  go test -v -timeout 30s -race -covermode=atomic -coverprofile=coverage.tmp -coverpkg "$COVERPKG" ./... 2>&1 | sed 's/ of statements in.*//; /warning: no packages being tested depend on matches for pattern /d'
  tail -n +2 coverage.tmp >> $COVERAGE

  rm coverage.tmp
  # Remove test only deps.
  go mod tidy
  popd
done
