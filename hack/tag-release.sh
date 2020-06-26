#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# This is run after the major release is published.

VERSION=v2.1.0

# It is intended that this file is run locally. For a full release tag, confirm the version is correct, and then:
#   ./hack/tag-release.sh --tag --push

CREATE_TAGS=0 # default is a dry run
PUSH_TAGS=0   # Assumes `upstream` is the remote name for sdk-go.

# Pick one:
REMOTE="origin"   # if checked out directly
#REMOTE="upstream" # if checked out with a fork

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

CE_SDK_V2=github.com/cloudevents/sdk-go/v2

echo --- All Modules ---
for gomodule in $(find . | grep "go\.mod" | awk '{gsub(/\/go.mod/,""); print $0}' | grep -v "./v2/test" | grep -v "./test")
do
  echo "  $gomodule"
  
  if [[ $gomodule = "./v2" ]]
  then
    echo "    skipping main module"
    continue
  fi
  
  pushd $gomodule > /dev/null
  
  repoint=$CE_SDK_V2
  
  if grep -Fq "$repoint" go.mod
  then
    tag="$VERSION"
    echo "    repointing dep on $CE_SDK_V2@$tag"
    go mod edit -dropreplace $repoint
    go get -d $repoint@$tag
    go mod tidy
  fi
  popd > /dev/null
    
done

echo --- Tagging ---

MODULES=(
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
    if [ "$CREATE_TAGS" -eq "1" ]; then
      echo "  tagging with $tag"
      git tag $tag
    fi
    if [ "$PUSH_TAGS" -eq "1" ]; then
      echo "  pushing $tag to $REMOTE"
      git push $REMOTE $tag
    fi
done
