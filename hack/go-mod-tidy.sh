#!/usr/bin/env bash

# Copyright 2022 The CloudEvents Authors
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

for gomodule in $(find . | grep "go\.mod" | awk '{gsub(/\/go.mod/,""); print $0}' | grep -v "./test" | grep -v "./conformance")
do
  pushd $gomodule
  go mod tidy -compat=1.17
  popd
done
