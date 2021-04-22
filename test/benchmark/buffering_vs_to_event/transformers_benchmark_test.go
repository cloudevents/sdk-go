/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package buffering_vs_to_event_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/buffering"
	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
	"github.com/cloudevents/sdk-go/v2/types"
)

var M binding.Message
var E *event.Event
var Err error

func BenchmarkBase(b *testing.B) {
	initialEvent := test.FullEvent()
	initialEvent.SetExtension("aaa", "bbb")
	for i := 0; i < b.N; i++ {
		M = bindingtest.MustCreateMockBinaryMessage(initialEvent)
	}
}

func BenchmarkToEventAndUpdateExtensions(b *testing.B) {
	initialEvent := test.FullEvent()
	ctx := context.TODO()
	initialEvent.SetExtension("aaa", "bbb")
	for i := 0; i < b.N; i++ {
		M = bindingtest.MustCreateMockBinaryMessage(initialEvent)
		E, _ = binding.ToEvent(ctx, M)
		if v, ok := E.Extensions()["aaa"]; ok {
			vStr, err := types.Format(v)
			if err != nil {
				panic(err)
			}
			E.SetExtension("aaa", strings.ToUpper(vStr))
		} else {
			E.SetExtension("aaa", strings.ToUpper("AAA"))
		}
		if v, ok := E.Extensions()["aTime"]; ok {
			vTime, err := types.ToTime(v)
			if err != nil {
				panic(err)
			}
			E.SetExtension("aTime", vTime.Add(3*time.Hour))
		} else {
			E.SetExtension("aTime", time.Now().UTC().Round(0))
		}

	}
}

func BenchmarkBufferingAndUpdateExtensions(b *testing.B) {
	initialEvent := test.FullEvent()
	ctx := context.TODO()
	initialEvent.SetExtension("aaa", "bbb")

	for i := 0; i < b.N; i++ {
		M = bindingtest.MustCreateMockBinaryMessage(initialEvent)
		M, _ = buffering.CopyMessage(ctx, M, transformers)
	}
}
