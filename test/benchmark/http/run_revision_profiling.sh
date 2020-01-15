#!/bin/bash

set -e

REVISION=$1

git checkout "$REVISION"
go build -v main.go

mkdir "$REVISION"

echo "Running baseline"
./main --bench=baseline -memprofile $REVISION/baseline_mem.pprof -cpuprofile $REVISION/baseline_cpu.pprof

echo "Running receiver sender"
./main --bench=receiver-sender -memprofile $REVISION/receiver_sender_mem.pprof -cpuprofile $REVISION/receiver_sender_cpu.pprof

echo "Running client"
./main --bench=client -memprofile $REVISION/client_mem.pprof -cpuprofile $REVISION/client_cpu.pprof

rm main
