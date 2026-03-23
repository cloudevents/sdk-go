/*
 Copyright 2026 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package extensions_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/extensions"
	"github.com/cloudevents/sdk-go/v2/test"
)

func TestAddExpiryTime(t *testing.T) {
	expiry := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	ext := extensions.ExpiryTimeExtension{ExpiryTime: expiry}

	e := event.New()
	e.SetSource("http://example.com/source")
	e.SetType("com.example.test")
	e.SetID("ABC-123")

	ext.AddExpiryTime(&e)

	got, ok := extensions.GetExpiryTime(e)
	require.True(t, ok)
	require.True(t, expiry.Equal(got.ExpiryTime))
}

func TestAddExpiryTime_Zero(t *testing.T) {
	ext := extensions.ExpiryTimeExtension{}

	e := event.New()
	e.SetSource("http://example.com/source")
	e.SetType("com.example.test")
	e.SetID("ABC-123")

	ext.AddExpiryTime(&e)

	_, ok := extensions.GetExpiryTime(e)
	require.False(t, ok)
}

func TestGetExpiryTime_NotSet(t *testing.T) {
	e := event.New()
	e.SetSource("http://example.com/source")
	e.SetType("com.example.test")
	e.SetID("ABC-123")

	_, ok := extensions.GetExpiryTime(e)
	require.False(t, ok)
}

func TestIsExpired(t *testing.T) {
	expiry := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	ext := extensions.ExpiryTimeExtension{ExpiryTime: expiry}

	e := event.New()
	e.SetSource("http://example.com/source")
	e.SetType("com.example.test")
	e.SetID("ABC-123")
	ext.AddExpiryTime(&e)

	require.True(t, extensions.IsExpired(e, expiry.Add(time.Hour)))
	require.False(t, extensions.IsExpired(e, expiry.Add(-time.Hour)))
}

func TestIsExpired_NoExtension(t *testing.T) {
	e := event.New()
	e.SetSource("http://example.com/source")
	e.SetType("com.example.test")
	e.SetID("ABC-123")

	require.False(t, extensions.IsExpired(e, time.Now()))
}

func TestExpiryTime_ReadTransformer_Empty(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	tests := []bindingtest.TransformerTestArgs{
		{
			Name:         "Read from Mock Structured message",
			InputMessage: bindingtest.MustCreateMockStructuredMessage(t, e),
			WantEvent:    e,
		},
		{
			Name:         "Read from Mock Binary message",
			InputMessage: bindingtest.MustCreateMockBinaryMessage(e),
			WantEvent:    e,
		},
		{
			Name:       "Read from Event message",
			InputEvent: e,
			WantEvent:  e,
		},
	}
	for _, tt := range tests {
		ext := extensions.ExpiryTimeExtension{}
		tt.Transformers = binding.Transformers{ext.ReadTransformer()}
		bindingtest.RunTransformerTests(t, context.TODO(), []bindingtest.TransformerTestArgs{tt})
		require.True(t, ext.ExpiryTime.IsZero())
	}
}

func TestExpiryTime_ReadTransformer(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()
	expiry := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	wantExt := extensions.ExpiryTimeExtension{ExpiryTime: expiry}
	wantExt.AddExpiryTime(&e)

	// Structured message mock round-trips through JSON, which converts
	// types.Timestamp to a plain string. Create a separate want event for it.
	eStructured := e.Clone()
	eStructured.SetExtension(extensions.ExpiryTimeExtensionKey, "2025-06-01T12:00:00Z")

	tests := []bindingtest.TransformerTestArgs{
		{
			Name:         "Read from Mock Structured message",
			InputMessage: bindingtest.MustCreateMockStructuredMessage(t, e),
			WantEvent:    eStructured,
		},
		{
			Name:         "Read from Mock Binary message",
			InputMessage: bindingtest.MustCreateMockBinaryMessage(e),
			WantEvent:    e,
		},
		{
			Name:       "Read from Event message",
			InputEvent: e,
			WantEvent:  e,
		},
	}
	for _, tt := range tests {
		haveExt := extensions.ExpiryTimeExtension{}
		tt.Transformers = binding.Transformers{haveExt.ReadTransformer()}
		bindingtest.RunTransformerTests(t, context.TODO(), []bindingtest.TransformerTestArgs{tt})
		require.True(t, expiry.Equal(haveExt.ExpiryTime))
	}
}

func TestExpiryTime_WriteTransformer(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()
	expiry := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	ext := extensions.ExpiryTimeExtension{ExpiryTime: expiry}
	want := e.Clone()
	ext.AddExpiryTime(&want)

	bindingtest.RunTransformerTests(t, context.TODO(), []bindingtest.TransformerTestArgs{
		{
			Name:         "Write to Mock Structured message",
			InputMessage: bindingtest.MustCreateMockStructuredMessage(t, e),
			WantEvent:    want,
			Transformers: binding.Transformers{ext.WriteTransformer()},
		},
		{
			Name:         "Write to Mock Binary message",
			InputMessage: bindingtest.MustCreateMockBinaryMessage(e),
			WantEvent:    want,
			Transformers: binding.Transformers{ext.WriteTransformer()},
		},
		{
			Name:         "Write to Event message",
			InputEvent:   e,
			WantEvent:    want,
			Transformers: binding.Transformers{ext.WriteTransformer()},
		},
	})
}
