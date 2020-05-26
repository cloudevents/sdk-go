#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

COVERAGE="`pwd`/coverage.txt"

# v2/test only
pushd ./test

# Run integration tests not in parallel
# Prepare coverage file only if not exists
if [ ! -f $COVERAGE ]; then
  touch ./coverage.tmp
  echo 'mode: atomic' > $COVERAGE
fi
COVERPKG=$(go list ./... | grep -v /vendor | tr "\n" ",")
for gomodule in $(go list ./integration/... | grep -v /cmd | grep -v /vendor)
do
  go test -v -parallel 1 -timeout 60s -race -covermode=atomic -coverprofile=coverage.tmp -coverpkg "$COVERPKG" "$gomodule" 2>&1 | sed 's/ of statements in.*//; /warning: no packages being tested depend on matches for pattern /d'
  tail -n +2 coverage.tmp >> $COVERAGE
done
rm coverage.tmp

# Remove test only deps.
go mod tidy

popd
