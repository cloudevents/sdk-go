package transformer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/types"

	. "github.com/cloudevents/sdk-go/pkg/binding/test"
)

func TestAddAttribute(t *testing.T) {
	e := MinEvent()
	e.Context = e.Context.AsV1()

	subject := "aaa"
	expectedEventWithSubject := e.Clone()
	require.NoError(t, expectedEventWithSubject.Context.SetSubject(subject))

	timestamp, err := types.ToTime(time.Now())
	require.NoError(t, err)
	expectedEventWithTime := e.Clone()
	require.NoError(t, expectedEventWithTime.Context.SetTime(timestamp))

	RunTransformerTests(t, context.Background(), []TransformerTestArgs{
		{
			Name:         "No change to id to Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(e.Clone()),
			WantEvent:    e.Clone(),
			Transformers: binding.TransformerFactories{AddAttribute(spec.ID, "new-id")},
		},
		{
			Name:         "No change to id to Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(e.Clone()),
			WantEvent:    e.Clone(),
			Transformers: binding.TransformerFactories{AddAttribute(spec.ID, "new-id")},
		},
		{
			Name:         "No change to id to Event message",
			InputEvent:   e,
			WantEvent:    CopyEventContext(e),
			Transformers: binding.TransformerFactories{AddAttribute(spec.ID, "new-id")},
		},
		{
			Name:         "Add subject to Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(e.Clone()),
			WantEvent:    expectedEventWithSubject,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Subject, subject)},
		},
		{
			Name:         "Add subject to Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(e.Clone()),
			WantEvent:    expectedEventWithSubject,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Subject, subject)},
		},
		{
			Name:         "Add subject to Event message",
			InputEvent:   e,
			WantEvent:    expectedEventWithSubject,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Subject, subject)},
		},
		{
			Name:         "Add time to Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(e.Clone()),
			WantEvent:    expectedEventWithTime,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Time, timestamp)},
		},
		{
			Name:         "Add time to Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(e.Clone()),
			WantEvent:    expectedEventWithTime,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Time, timestamp)},
		},
		{
			Name:         "Add time to Event message",
			InputEvent:   e,
			WantEvent:    expectedEventWithTime,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Time, timestamp)},
		},
	})
}

func TestAddExtension(t *testing.T) {
	e := MinEvent()
	e.Context = e.Context.AsV1()

	extName := "aaa"
	extValue := "bbb"
	expectedEventWithExtension := e.Clone()
	require.NoError(t, expectedEventWithExtension.Context.SetExtension(extName, extValue))

	RunTransformerTests(t, context.Background(), []TransformerTestArgs{
		{
			Name:         "No change to extension 'aaa' to Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(expectedEventWithExtension.Clone()),
			WantEvent:    expectedEventWithExtension.Clone(),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "No change to extension 'aaa' to Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(expectedEventWithExtension.Clone()),
			WantEvent:    expectedEventWithExtension.Clone(),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "No change to extension 'aaa' to Event message",
			InputEvent:   expectedEventWithExtension,
			WantEvent:    CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "Add extension 'aaa' to Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(e.Clone()),
			WantEvent:    expectedEventWithExtension.Clone(),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "Add extension 'aaa' to Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(e.Clone()),
			WantEvent:    expectedEventWithExtension.Clone(),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "Add extension 'aaa' to Event message",
			InputEvent:   e,
			WantEvent:    CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
	})
}
