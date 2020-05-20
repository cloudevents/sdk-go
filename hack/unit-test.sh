#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

for amod in $(find . | grep "go\.mod" | awk '{gsub(/\/go.mod/,""); print $0}' | grep -v "./v2/test")

do
  echo --- Testing $amod ---
  pushd $amod
  touch ./coverage.tmp
  echo 'mode: atomic' > ./coverage.txt
  COVERPKG=$(go list ./... | grep -v /vendor | grep -v /test | tr "\n" ",")

  go test -v -timeout 15s -race -covermode=atomic -coverprofile=coverage.tmp -coverpkg "$COVERPKG" ./... 2>&1 | sed 's/ of statements in.*//; /warning: no packages being tested depend on matches for pattern /d'
  tail -n +2 coverage.tmp >> ./coverage.txt

  rm coverage.tmp
  # Remove test only deps.
  go mod tidy
  popd
done
