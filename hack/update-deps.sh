#!/bin/bash
# Copyright 2022 The CloudEvents Authors
# SPDX-License-Identifier: Apache-2.0
# update-deps.sh - Updates Go dependencies in all directories with go.mod files
#
# This script:
# 1. Finds all directories containing go.mod files
# 2. Goes into each directory and runs go get -u to update dependencies
# 3. Runs go mod tidy to clean up the go.mod and go.sum files

set -euo pipefail

echo "====================================="
echo "Go Dependencies Update Script"
echo "====================================="

echo "Finding all directories with go.mod files..."
DIRS=$(find . -name "go.mod" -exec dirname {} \; | sort)
if [ -z "$DIRS" ]; then
  echo "No go.mod files found!"
  exit 0
fi

DIR_COUNT=$(echo "$DIRS" | wc -l | tr -d ' ')
echo "Found $DIR_COUNT directories with go.mod files"
echo

COUNTER=1
for DIR in $DIRS; do
  echo "[$COUNTER/$DIR_COUNT] Processing $DIR"

  pushd "$DIR" >/dev/null

  echo "  - Updating dependencies..."
  go get -u -t ./...

  echo "  - Running go mod tidy..."
  go mod tidy

  popd >/dev/null

  echo "  - Done"
  echo

  COUNTER=$((COUNTER + 1))
done

# Satellite modules `replace` github.com/cloudevents/sdk-go/v2 with ../../../v2, so their
# tidy result depends on v2/go.mod as it is on disk at that moment. A module tidied before
# v2 gets its own `go get -u` misses any go directive bump v2 picks up later, and CI then
# fails with "updates to go.mod needed". No single ordering fixes this, so tidy everything
# once more now that every go.mod has its final dependencies.
echo "====================================="
echo "Re-tidying all modules"
echo "====================================="
echo

COUNTER=1
for DIR in $DIRS; do
  echo "[$COUNTER/$DIR_COUNT] Re-tidying $DIR"
  pushd "$DIR" >/dev/null
  go mod tidy
  popd >/dev/null
  COUNTER=$((COUNTER + 1))
done
echo

echo "All dependencies updated successfully!"
