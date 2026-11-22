/*
 Copyright 2026 The CloudEvents Authors
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

func TestAddSampledRate(t *testing.T) {
	ext := extensions.SampledRateExtension{SampledRate: 30}

	e := event.New()
	e.SetSource("http://example.com/source")
	e.SetType("com.example.test")
	e.SetID("ABC-123")

	require.NoError(t, ext.AddSampledRate(&e))

	got, ok := extensions.GetSampledRate(e)
	require.True(t, ok)
	require.Equal(t, int32(30), got.SampledRate)
}

func TestAddSampledRate_Zero(t *testing.T) {
	e := event.New()
	require.NoError(t, extensions.SampledRateExtension{}.AddSampledRate(&e))

	_, ok := extensions.GetSampledRate(e)
	require.False(t, ok)
}

func TestAddSampledRate_Negative(t *testing.T) {
	e := event.New()
	require.Error(t, extensions.SampledRateExtension{SampledRate: -1}.AddSampledRate(&e))

	_, ok := extensions.GetSampledRate(e)
	require.False(t, ok)
}

func TestGetSampledRate_NotSet(t *testing.T) {
	_, ok := extensions.GetSampledRate(event.New())
	require.False(t, ok)
}

func TestGetSampledRate_Invalid(t *testing.T) {
	for name, v := range map[string]interface{}{"zero": int32(0), "negative": int32(-5), "not a number": "abc"} {
		t.Run(name, func(t *testing.T) {
			e := event.New()
			e.SetExtension(extensions.SampledRateExtensionKey, v)
			_, ok := extensions.GetSampledRate(e)
			require.False(t, ok)
		})
	}
}

func TestSampledRate_ReadTransformer_Empty(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	for _, tt := range []bindingtest.TransformerTestArgs{
		{Name: "Read from Mock Structured message", InputMessage: bindingtest.MustCreateMockStructuredMessage(t, e), WantEvent: e},
		{Name: "Read from Mock Binary message", InputMessage: bindingtest.MustCreateMockBinaryMessage(e), WantEvent: e},
		{Name: "Read from Event message", InputEvent: e, WantEvent: e},
	} {
		ext := extensions.SampledRateExtension{}
		tt.Transformers = binding.Transformers{ext.ReadTransformer()}
		bindingtest.RunTransformerTests(t, context.TODO(), []bindingtest.TransformerTestArgs{tt})
		require.Zero(t, ext.SampledRate)
	}
}

func TestSampledRate_ReadTransformer(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()
	require.NoError(t, extensions.SampledRateExtension{SampledRate: 100}.AddSampledRate(&e))

	// The structured mock round-trips through JSON, which turns the integer into a float64.
	eStructured := e.Clone()
	eStructured.SetExtension(extensions.SampledRateExtensionKey, float64(100))

	for _, tt := range []bindingtest.TransformerTestArgs{
		{Name: "Read from Mock Structured message", InputMessage: bindingtest.MustCreateMockStructuredMessage(t, e), WantEvent: eStructured},
		{Name: "Read from Mock Binary message", InputMessage: bindingtest.MustCreateMockBinaryMessage(e), WantEvent: e},
		{Name: "Read from Event message", InputEvent: e, WantEvent: e},
	} {
		have := extensions.SampledRateExtension{}
		tt.Transformers = binding.Transformers{have.ReadTransformer()}
		bindingtest.RunTransformerTests(t, context.TODO(), []bindingtest.TransformerTestArgs{tt})
		require.Equal(t, int32(100), have.SampledRate)
	}
}

func TestSampledRate_ReadTransformer_Invalid(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()
	e.SetExtension(extensions.SampledRateExtensionKey, int32(0))

	ext := extensions.SampledRateExtension{}
	m := bindingtest.MustCreateMockBinaryMessage(e).(*bindingtest.MockBinaryMessage)
	require.Error(t, ext.ReadTransformer()(m, m))
}

func TestSampledRate_WriteTransformer(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	ext := extensions.SampledRateExtension{SampledRate: 30}
	want := e.Clone()
	require.NoError(t, ext.AddSampledRate(&want))

	bindingtest.RunTransformerTests(t, context.TODO(), []bindingtest.TransformerTestArgs{
		{Name: "Write to Mock Structured message", InputMessage: bindingtest.MustCreateMockStructuredMessage(t, e), WantEvent: want, Transformers: binding.Transformers{ext.WriteTransformer()}},
		{Name: "Write to Mock Binary message", InputMessage: bindingtest.MustCreateMockBinaryMessage(e), WantEvent: want, Transformers: binding.Transformers{ext.WriteTransformer()}},
		{Name: "Write to Event message", InputEvent: e, WantEvent: want, Transformers: binding.Transformers{ext.WriteTransformer()}},
	})
}

func TestSampledRate_WriteTransformer_Negative(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()
	ext := extensions.SampledRateExtension{SampledRate: -1}
	m := bindingtest.MustCreateMockBinaryMessage(e).(*bindingtest.MockBinaryMessage)
	require.Error(t, ext.WriteTransformer()(m, m))
}
