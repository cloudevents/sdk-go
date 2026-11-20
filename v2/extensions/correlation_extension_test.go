/*
 Copyright 2021 The CloudEvents Authors
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

func TestCorrelationExtension(t *testing.T) {
	testCases := []struct {
		name          string
		extension     extensions.CorrelationExtension
		expectedEvent func() event.Event
	}{
		{
			name: "both attributes",
			extension: extensions.CorrelationExtension{
				CorrelationID: "corr-1",
				CausationID:   "caus-1",
			},
			expectedEvent: func() event.Event {
				e := test.MinEvent()
				e.SetExtension("correlationid", "corr-1")
				e.SetExtension("causationid", "caus-1")
				return e
			},
		},
		{
			name: "only correlationid",
			extension: extensions.CorrelationExtension{
				CorrelationID: "corr-1",
			},
			expectedEvent: func() event.Event {
				e := test.MinEvent()
				e.SetExtension("correlationid", "corr-1")
				return e
			},
		},
		{
			name: "only causationid",
			extension: extensions.CorrelationExtension{
				CausationID: "caus-1",
			},
			expectedEvent: func() event.Event {
				e := test.MinEvent()
				e.SetExtension("causationid", "caus-1")
				return e
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := test.MinEvent()
			tc.extension.AddCorrelationAttributes(&e)

			expected := tc.expectedEvent()
			require.Equal(t, expected.Extensions(), e.Extensions())

			// Test GetCorrelationExtension
			got, ok := extensions.GetCorrelationExtension(e)
			require.True(t, ok)
			require.Equal(t, tc.extension, got)
		})
	}
}

func TestCorrelationExtension_GetNotSet(t *testing.T) {
	e := test.MinEvent()
	_, ok := extensions.GetCorrelationExtension(e)
	require.False(t, ok)
}

func TestCorrelationExtension_ReadTransformer(t *testing.T) {
	e := test.MinEvent()
	e.SetExtension("correlationid", "corr-1")
	e.SetExtension("causationid", "caus-1")

	ext := extensions.CorrelationExtension{}
	bindingtest.RunTransformerTests(t, context.TODO(), []bindingtest.TransformerTestArgs{
		{
			Name:         "Read from Mock Structured message",
			InputMessage: bindingtest.MustCreateMockStructuredMessage(t, e),
			WantEvent:    e,
			Transformers: binding.Transformers{ext.ReadTransformer()},
		},
	})
	require.Equal(t, "corr-1", ext.CorrelationID)
	require.Equal(t, "caus-1", ext.CausationID)
}

func TestCorrelationExtension_WriteTransformer(t *testing.T) {
	e := test.MinEvent()
	ext := extensions.CorrelationExtension{
		CorrelationID: "corr-1",
		CausationID:   "caus-1",
	}

	want := e.Clone()
	want.SetExtension("correlationid", "corr-1")
	want.SetExtension("causationid", "caus-1")

	bindingtest.RunTransformerTests(t, context.TODO(), []bindingtest.TransformerTestArgs{
		{
			Name:         "Write to Mock Structured message",
			InputMessage: bindingtest.MustCreateMockStructuredMessage(t, e),
			WantEvent:    want,
			Transformers: binding.Transformers{ext.WriteTransformer()},
		},
	})
}
