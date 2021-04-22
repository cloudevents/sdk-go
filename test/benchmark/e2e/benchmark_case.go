/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package e2e

type BenchmarkCase struct {
	PayloadSize   int
	Parallelism   int
	OutputSenders int
}

func GenerateAllBenchmarkCases(
	payloadMin int,
	payloadMax int,
	parallelismMin int,
	parallelismMax int,
	outputSendersMin int,
	outputSendersMax int,
) []BenchmarkCase {
	var cases []BenchmarkCase

	for payload := payloadMin; payload <= payloadMax; payload *= 2 {
		for parallelism := parallelismMin; parallelism <= parallelismMax; parallelism += 1 {
			for outputSenders := outputSendersMin; outputSenders <= outputSendersMax; outputSenders += 1 {
				cases = append(cases, BenchmarkCase{payload, parallelism, outputSenders})
			}
		}
	}

	return cases
}
