/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package e2e

import (
	"encoding/csv"
	"strconv"
	"testing"
)

type BenchmarkResult struct {
	BenchmarkCase
	testing.BenchmarkResult
}

func (br *BenchmarkResult) record() []string {
	return []string{
		strconv.Itoa(br.Parallelism),
		strconv.Itoa(br.PayloadSize),
		strconv.Itoa(br.OutputSenders),
		strconv.FormatInt(br.NsPerOp(), 10),
		strconv.FormatInt(br.AllocedBytesPerOp(), 10),
	}
}

type BenchmarkResults []BenchmarkResult

func (br BenchmarkResults) WriteToCsv(writer *csv.Writer) error {
	for _, i2 := range br {
		err := writer.Write(i2.record())
		if err != nil {
			return err
		}
	}
	return nil
}
