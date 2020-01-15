#!/bin/bash

set -e

REVISION=$1

git checkout "$REVISION"
go build -v main.go

mkdir "$REVISION"

echo "Running baseline"
./main --bench=baseline > "$REVISION/baseline.csv"

echo "Running receiver sender"
./main --bench=receiver-sender > "$REVISION/receiver-sender.csv"

echo "Running client"
./main --bench=client > "$REVISION/client.csv"

rm main