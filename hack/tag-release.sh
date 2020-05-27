#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

echo --- All modules in this project:
for gomodule in $(find . | grep "go\.mod" | awk '{gsub(/\/go.mod/,""); print $0}' | grep -v "./v2/test")
do
  echo "   - $gomodule"
done
