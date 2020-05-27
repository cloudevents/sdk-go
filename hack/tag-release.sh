#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

VERSION=v2.0.0

# It is intended that this file is run locally. For a full release tag, confirm the version is correct, and then:
#   ./hack/tag-release.sh --tag --push

CREATE_TAGS=0 # default is a dry run
PUSH_TAGS=0   # Assumes `upstream` is the remote name for sdk-go.

# Loop through arguments and process them
for arg in "$@"
do
    case $arg in
        -t|--tag)
        CREATE_TAGS=1
        shift
        ;;
        -p|--push)
        PUSH_TAGS=1
        shift
        ;;
    esac
done

echo --- All Modules ---
for gomodule in $(find . | grep "go\.mod" | awk '{gsub(/\/go.mod/,""); print $0}' | grep -v "./v2/test")
do
  echo "  $gomodule"
done

echo --- Tagging ---

MODULES=(
  ""               # root module
  "protocol/amqp"
  "protocol/stan"
  "protocol/nats"
  "protocol/pubsub"
  "protocol/kafka_sarama"
)

for i in "${MODULES[@]}"; do
    tag=""
    if [ "$i" = "" ]; then
      tag="$VERSION"
    else
      tag="$i/$VERSION"
    fi
    echo "  $tag"
    if [ "$CREATE_TAGS" -eq "1" ]; then
      git tag $TAG
    fi
    if [ "$PUSH_TAGS" -eq "1" ]; then
      git push upstream $TAG
    fi
done
