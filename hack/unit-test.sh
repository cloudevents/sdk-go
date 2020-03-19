#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# v1
pushd ./v1
touch ./coverage.tmp
echo 'mode: atomic' > ./coverage.txt
COVERPKG=$(go list ./... | grep -v /vendor | tr "\n" ",")
for gomodule in $(go list ./... | grep -v /cmd | grep -v /vendor)
do
  go test -v -timeout 15s -covermode=atomic -coverprofile=coverage.tmp -coverpkg "$COVERPKG" "$gomodule" 2>&1 | sed 's/ of statements in.*//; /warning: no packages being tested depend on matches for pattern /d'
  tail -n +2 coverage.tmp >> ./coverage.txt
done
rm coverage.tmp
# Remove test only deps.
go mod tidy
popd

# v2
pushd ./v2
touch ./coverage.tmp
echo 'mode: atomic' > ./coverage.txt
COVERPKG=$(go list ./... | grep -v /vendor | grep -v /test | tr "\n" ",")
for gomodule in $(go list ./... | grep -v /cmd | grep -v /vendor | grep -v /test)
do
  go test -v -timeout 15s -covermode=atomic -coverprofile=coverage.tmp -coverpkg "$COVERPKG" "$gomodule" 2>&1 | sed 's/ of statements in.*//; /warning: no packages being tested depend on matches for pattern /d'
  tail -n +2 coverage.tmp >> ./coverage.txt
done
rm coverage.tmp
# Remove test only deps.
go mod tidy
popd