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
DIRS=$(find . -name "go.mod" -exec dirname {} \;)
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

echo "All dependencies updated successfully!"
