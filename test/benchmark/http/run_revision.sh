#!/bin/bash

set -e

RUNNABLE_NAME="http-bench"

function usage {
   echo "Usage: $0 [--cpu-profile] [--mem-profile] [--max-parallelism n] [--max-payload n] [--max-output-senders n] [git_revision]"
   exit 1
}

ADDITIONAL_ARGS=""

PARAMS=""
while (( "$#" )); do
  case "$1" in
    -h|--help)
      usage
      ;;
    -c|--cpu-profile)
      CPU_PROFILE="1"
      shift
      ;;
    -m|--mem-profile)
      MEM_PROFILE="1"
      shift
      ;;
    --max-parallelism)
      ADDITIONAL_ARGS="$ADDITIONAL_ARGS --max-parallelism $2"
      shift 2
      ;;
    --max-payload)
      ADDITIONAL_ARGS="$ADDITIONAL_ARGS --max-payload $2"
      shift 2
      ;;
    --max-output-senders)
      ADDITIONAL_ARGS="$ADDITIONAL_ARGS --max-output-senders $2"
      shift 2
      ;;
    --) # end argument parsing
      shift
      break
      ;;
    -*|--*=) # unsupported flags
      echo "Error: Unsupported flag $1" >&2
      exit 1
      ;;
    *) # preserve positional arguments
      PARAMS="$PARAMS $1"
      shift
      ;;
  esac
done
eval set -- "$PARAMS"

REVISION=$1
if [ ! -z "$REVISION" ]
then
      git checkout "$REVISION"
else
      REVISION=results
fi

go build -o $RUNNABLE_NAME -v github.com/cloudevents/sdk-go/test/benchmark/http

mkdir -p "$REVISION"

BENCHS=(
  "baseline-structured"
  "baseline-binary"
  "binding-structured-to-structured"
  "binding-structured-to-binary"
  "binding-binary-to-structured"
  "binding-binary-to-binary"
  "client-binary"
  "client-structured"
)

for i in "${BENCHS[@]}"; do
    ./$RUNNABLE_NAME --bench="$i" \
    ${MEM_PROFILE+-memprofile $REVISION/$i-mem.pprof} \
    ${CPU_PROFILE+-cpuprofile $REVISION/$i-cpu.pprof} \
    $ADDITIONAL_ARGS \
    --out="$REVISION/$i.csv"
done

rm $RUNNABLE_NAME
