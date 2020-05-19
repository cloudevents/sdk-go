#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# v2 only
pushd ./v2

# Run integration tests not in parallel
# Prepare coverage file only if not exists
if [ ! -f ./coverage.txt ]; then
  touch ./coverage.tmp
  echo 'mode: atomic' > ./coverage.txt
fi
COVERPKG=$(go list ./... | grep -v /vendor | tr "\n" ",")
for gomodule in $(go list ./test/integration/... | grep -v /cmd | grep -v /vendor)
do
  go test -v -parallel 1 -timeout 60s -race -covermode=atomic -coverprofile=coverage.tmp -coverpkg "$COVERPKG" "$gomodule" 2>&1 | sed 's/ of statements in.*//; /warning: no packages being tested depend on matches for pattern /d'
  tail -n +2 coverage.tmp >> ./coverage.txt
done

# special case pubsub
go test -v -tags=integration -parallel 1 -timeout 60s -race -covermode=atomic -coverprofile=coverage.tmp -coverpkg "$COVERPKG" "./protocol/pubsub/internal" 2>&1 | sed 's/ of statements in.*//; /warning: no packages being tested depend on matches for pattern /d'
tail -n +2 coverage.tmp >> ./coverage.txt

rm coverage.tmp

# Remove test only deps.
go mod tidy

popd
