#!/usr/bin/env bash

# Copyright 2026 The CloudEvents Authors
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

GOVULNCHECK_VERSION="v1.9.0"
FAILED=()

for gomodule in $(find . -name go.mod -exec dirname {} \; | sort)
do
  echo
  echo --- Checking $gomodule ---
  echo

  pushd $gomodule >/dev/null
  go run golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION} ./... || FAILED+=("$gomodule")
  popd >/dev/null
done

if [ ${#FAILED[@]} -ne 0 ]; then
  echo
  echo "govulncheck found vulnerabilities in: ${FAILED[*]}"
  exit 1
fi
