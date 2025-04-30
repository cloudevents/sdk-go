/*
 Copyright 2023 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package json

import (
	"encoding/json"
	"testing"

	goccyjson "github.com/goccy/go-json"
	jsoniter "github.com/json-iterator/go"
)

type testStruct struct {
	ID    string `json:"id"`
	Value string `json:"value"`
	Count int    `json:"count"`
	Items []item `json:"items"`
}

type item struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

var testData = testStruct{
	ID:    "test-id",
	Value: "test-value",
	Count: 42,
	Items: []item{
		{Name: "item1", Amount: 100},
		{Name: "item2", Amount: 200},
		{Name: "item3", Amount: 300},
	},
}

func BenchmarkMarshal(b *testing.B) {
	b.Run("encoding/json", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = json.Marshal(testData)
		}
	})

	b.Run("jsoniter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = jsoniter.Marshal(testData)
		}
	})

	b.Run("goccy/go-json", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = goccyjson.Marshal(testData)
		}
	})
}

func BenchmarkUnmarshal(b *testing.B) {
	standardJSON, _ := json.Marshal(testData)
	jsoniterData, _ := jsoniter.Marshal(testData)
	goccyData, _ := goccyjson.Marshal(testData)

	b.Run("encoding/json", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result testStruct
			_ = json.Unmarshal(standardJSON, &result)
		}
	})

	b.Run("jsoniter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result testStruct
			_ = jsoniter.Unmarshal(jsoniterData, &result)
		}
	})

	b.Run("goccy/go-json", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result testStruct
			_ = goccyjson.Unmarshal(goccyData, &result)
		}
	})
}
