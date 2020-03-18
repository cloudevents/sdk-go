package transcoder

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/spec"
	"github.com/cloudevents/sdk-go/v1/binding/test"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
)

func TestAddAttribute(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	subject := "aaa"
	expectedEventWithSubject := test.CopyEventContext(e)
	require.NoError(t, expectedEventWithSubject.Context.SetSubject(subject))

	timestamp, err := types.ToTime(time.Now())
	require.NoError(t, err)
	expectedEventWithTime := test.CopyEventContext(e)
	require.NoError(t, expectedEventWithTime.Context.SetTime(timestamp))

	test.RunTranscoderTests(t, context.Background(), []test.TranscoderTestArgs{
		{
			Name:         "No change to id to Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(e)),
			WantEvent:    test.CopyEventContext(e),
			Transformers: binding.TransformerFactories{AddAttribute(spec.ID, "new-id")},
		},
		{
			Name:         "No change to id to Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(e)),
			WantEvent:    test.CopyEventContext(e),
			Transformers: binding.TransformerFactories{AddAttribute(spec.ID, "new-id")},
		},
		{
			Name:         "No change to id to Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(e)),
			WantEvent:    test.CopyEventContext(e),
			Transformers: binding.TransformerFactories{AddAttribute(spec.ID, "new-id")},
		},
		{
			Name:         "Add subject to Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(e)),
			WantEvent:    expectedEventWithSubject,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Subject, subject)},
		},
		{
			Name:         "Add subject to Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(e)),
			WantEvent:    expectedEventWithSubject,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Subject, subject)},
		},
		{
			Name:         "Add subject to Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(e)),
			WantEvent:    expectedEventWithSubject,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Subject, subject)},
		},
		{
			Name:         "Add time to Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(e)),
			WantEvent:    expectedEventWithTime,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Time, timestamp)},
		},
		{
			Name:         "Add time to Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(e)),
			WantEvent:    expectedEventWithTime,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Time, timestamp)},
		},
		{
			Name:         "Add time to Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(e)),
			WantEvent:    expectedEventWithTime,
			Transformers: binding.TransformerFactories{AddAttribute(spec.Time, timestamp)},
		},
	})
}

func TestAddExtension(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	extName := "aaa"
	extValue := "bbb"
	expectedEventWithExtension := test.CopyEventContext(e)
	require.NoError(t, expectedEventWithExtension.Context.SetExtension(extName, extValue))

	test.RunTranscoderTests(t, context.Background(), []test.TranscoderTestArgs{
		{
			Name:         "No change to extension 'aaa' to Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(expectedEventWithExtension)),
			WantEvent:    test.CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "No change to extension 'aaa' to Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(expectedEventWithExtension)),
			WantEvent:    test.CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "No change to extension 'aaa' to Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(expectedEventWithExtension)),
			WantEvent:    test.CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "Add extension 'aaa' to Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(e)),
			WantEvent:    test.CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "Add extension 'aaa' to Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(e)),
			WantEvent:    test.CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			Name:         "Add extension 'aaa' to Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(e)),
			WantEvent:    test.CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{AddExtension(extName, extValue)},
		},
	})
}
