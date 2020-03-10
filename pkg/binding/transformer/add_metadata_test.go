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
	expectedEventWithSubject := CopyEventContext(e)
	require.NoError(t, expectedEventWithSubject.Context.SetSubject(subject))

	timestamp, err := types.ToTime(time.Now())
	require.NoError(t, err)
	expectedEventWithTime := CopyEventContext(e)
	require.NoError(t, expectedEventWithTime.Context.SetTime(timestamp))

	RunTransformerTests(t, context.Background(), []TransformerTestArgs{
		{
			Name:         "No change to id to Mock Structured message",
			InputMessage: MustCreateMockStructuredMessage(CopyEventContext(e)),
			WantEvent:    CopyEventContext(e),
			Transformers: binding.TransformerFactories{AddAttribute(spec.ID, "new-id")},
		},
		{
			Name:         "No change to id to Mock Binary message",
			InputMessage: MustCreateMockBinaryMessage(CopyEventContext(e)),
			WantEvent:    CopyEventContext(e),
			Transformers: binding.TransformerFactories{AddAttribute(spec.ID, "new-id")},
		},
		{
			Name:         "No change to id to Event message",
			InputMessage: binding.EventMessage(CopyEventContext(e)),
			WantEvent:    CopyEventContext(e),
			Transformers: binding.TransformerFactories{AddAttribute(spec.ID, "new-id")},
		},
		{
			Name:         "Add subject to Mock Structured message",
			InputMessage: MustCreateMockStructuredMessage(CopyEventContext(e)),
			WantEvent:    expectedEventWithSubject,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Subject, subject)},
		},
		{
			Name:         "Add subject to Mock Binary message",
			InputMessage: MustCreateMockBinaryMessage(CopyEventContext(e)),
			WantEvent:    expectedEventWithSubject,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Subject, subject)},
		},
		{
			Name:         "Add subject to Event message",
			InputMessage: binding.EventMessage(CopyEventContext(e)),
			WantEvent:    expectedEventWithSubject,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Subject, subject)},
		},
		{
			Name:         "Add time to Mock Structured message",
			InputMessage: MustCreateMockStructuredMessage(CopyEventContext(e)),
			WantEvent:    expectedEventWithTime,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Time, timestamp)},
		},
		{
			Name:         "Add time to Mock Binary message",
			InputMessage: MustCreateMockBinaryMessage(CopyEventContext(e)),
			WantEvent:    expectedEventWithTime,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Time, timestamp)},
		},
		{
			Name:         "Add time to Event message",
			InputMessage: binding.EventMessage(CopyEventContext(e)),
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
	expectedEventWithExtension := CopyEventContext(e)
	require.NoError(t, expectedEventWithExtension.Context.SetExtension(extName, extValue))

	RunTransformerTests(t, context.Background(), []TransformerTestArgs{
		{
			Name:         "No change to extension 'aaa' to Mock Structured message",
			InputMessage: MustCreateMockStructuredMessage(CopyEventContext(expectedEventWithExtension)),
			WantEvent:    CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "No change to extension 'aaa' to Mock Binary message",
			InputMessage: MustCreateMockBinaryMessage(CopyEventContext(expectedEventWithExtension)),
			WantEvent:    CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "No change to extension 'aaa' to Event message",
			InputMessage: binding.EventMessage(CopyEventContext(expectedEventWithExtension)),
			WantEvent:    CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "Add extension 'aaa' to Mock Structured message",
			InputMessage: MustCreateMockStructuredMessage(CopyEventContext(e)),
			WantEvent:    CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "Add extension 'aaa' to Mock Binary message",
			InputMessage: MustCreateMockBinaryMessage(CopyEventContext(e)),
			WantEvent:    CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "Add extension 'aaa' to Event message",
			InputMessage: binding.EventMessage(CopyEventContext(e)),
			WantEvent:    CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
	})
}
