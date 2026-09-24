/*
 Copyright 2024 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package extensions_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/extensions"
	"github.com/cloudevents/sdk-go/v2/test"
)

func TestAddSequenceExtension(t *testing.T) {
	e := event.New()
	e.SetSource("https://example.com/source")
	e.SetType("com.example.test")
	e.SetID("123")

	extensions.AddSequenceExtension(&e, "001")

	got, ok := e.Extensions()[extensions.SequenceExtensionKey]
	require.True(t, ok)
	require.Equal(t, "001", got)
}

func TestAddSequenceExtension_Empty(t *testing.T) {
	e := event.New()
	e.SetSource("https://example.com/source")
	e.SetType("com.example.test")
	e.SetID("123")

	extensions.AddSequenceExtension(&e, "")

	_, ok := e.Extensions()[extensions.SequenceExtensionKey]
	require.False(t, ok)
}

func TestGetSequenceExtension(t *testing.T) {
	e := event.New()
	e.SetSource("https://example.com/source")
	e.SetType("com.example.test")
	e.SetID("123")
	e.SetExtension(extensions.SequenceExtensionKey, "042")

	ext, ok := extensions.GetSequenceExtension(e)
	require.True(t, ok)
	require.Equal(t, "042", ext.Sequence)
}

func TestGetSequenceExtension_NotPresent(t *testing.T) {
	e := event.New()
	e.SetSource("https://example.com/source")
	e.SetType("com.example.test")
	e.SetID("123")

	ext, ok := extensions.GetSequenceExtension(e)
	require.False(t, ok)
	require.Equal(t, "", ext.Sequence)
}

func TestSequenceExtension_ReadTransformer_Empty(t *testing.T) {
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
		ext := extensions.SequenceExtension{}
		tt.Transformers = binding.Transformers{ext.ReadTransformer()}
		bindingtest.RunTransformerTests(t, context.TODO(), []bindingtest.TransformerTestArgs{tt})
		require.Zero(t, ext.Sequence)
	}
}

func TestSequenceExtension_ReadTransformer(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()
	wantExt := extensions.SequenceExtension{
		Sequence: "00042",
	}
	extensions.AddSequenceExtension(&e, wantExt.Sequence)

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
		haveExt := extensions.SequenceExtension{}
		tt.Transformers = binding.Transformers{haveExt.ReadTransformer()}
		bindingtest.RunTransformerTests(t, context.TODO(), []bindingtest.TransformerTestArgs{tt})
		require.Equal(t, wantExt.Sequence, haveExt.Sequence)
	}
}

func TestSequenceExtension_WriteTransformer(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	ext := extensions.SequenceExtension{
		Sequence: "00042",
	}
	want := e.Clone()
	extensions.AddSequenceExtension(&want, ext.Sequence)

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
