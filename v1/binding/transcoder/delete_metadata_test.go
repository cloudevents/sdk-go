package transcoder

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/spec"
	"github.com/cloudevents/sdk-go/v1/binding/test"
)

func TestDeleteAttribute(t *testing.T) {
	withSubjectEvent := test.MinEvent()
	withSubjectEvent.Context = withSubjectEvent.Context.AsV1()
	require.NoError(t, withSubjectEvent.Context.SetSubject("aaa"))

	withTimeEvent := test.CopyEventContext(withSubjectEvent)
	require.NoError(t, withTimeEvent.Context.SetTime(time.Now()))

	noSubjectEvent := test.CopyEventContext(withSubjectEvent)
	require.NoError(t, noSubjectEvent.Context.SetSubject(""))

	test.RunTranscoderTests(t, context.Background(), []test.TranscoderTestArgs{
		{
			Name:         "Remove subject from Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(withSubjectEvent)),
			WantEvent:    noSubjectEvent,
			Transformers: binding.TransformerFactories{DeleteAttribute(spec.Subject)},
		},
		{
			Name:         "Remove subject from Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(withSubjectEvent)),
			WantEvent:    noSubjectEvent,
			Transformers: binding.TransformerFactories{DeleteAttribute(spec.Subject)},
		},
		{
			Name:         "Remove subject from Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(withSubjectEvent)),
			WantEvent:    noSubjectEvent,
			Transformers: binding.TransformerFactories{DeleteAttribute(spec.Subject)},
		},
		{
			Name:         "Remove time from Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(withTimeEvent)),
			WantEvent:    test.CopyEventContext(withSubjectEvent),
			Transformers: binding.TransformerFactories{DeleteAttribute(spec.Time)},
		},
		{
			Name:         "Remove time from Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(withTimeEvent)),
			WantEvent:    test.CopyEventContext(withSubjectEvent),
			Transformers: binding.TransformerFactories{DeleteAttribute(spec.Time)},
		},
		{
			Name:         "Remove time from Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(withTimeEvent)),
			WantEvent:    test.CopyEventContext(withSubjectEvent),
			Transformers: binding.TransformerFactories{DeleteAttribute(spec.Time)},
		},
		{
			Name:         "Do nothing with Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(withSubjectEvent)),
			WantEvent:    test.CopyEventContext(withSubjectEvent),
			Transformers: binding.TransformerFactories{DeleteAttribute(spec.Time)},
		},
		{
			Name:         "Do nothing with Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(withSubjectEvent)),
			WantEvent:    test.CopyEventContext(withSubjectEvent),
			Transformers: binding.TransformerFactories{DeleteAttribute(spec.Time)},
		},
		{
			Name:         "Do nothing with Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(withSubjectEvent)),
			WantEvent:    test.CopyEventContext(withSubjectEvent),
			Transformers: binding.TransformerFactories{DeleteAttribute(spec.Time)},
		},
	})
}

func TestDeleteExtension(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	extName := "aaa"
	extValue := "bbb"
	expectedEventWithExtension := test.CopyEventContext(e)
	require.NoError(t, expectedEventWithExtension.Context.SetExtension(extName, extValue))

	test.RunTranscoderTests(t, context.Background(), []test.TranscoderTestArgs{
		{
			Name:         "No change to Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(expectedEventWithExtension)),
			WantEvent:    test.CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{DeleteExtension("ccc")},
		},
		{
			Name:         "No change to Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(expectedEventWithExtension)),
			WantEvent:    test.CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{DeleteExtension("ccc")},
		},
		{
			Name:         "No change to Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(expectedEventWithExtension)),
			WantEvent:    test.CopyEventContext(expectedEventWithExtension),
			Transformers: binding.TransformerFactories{DeleteExtension("ccc")},
		},
		{
			Name:         "Delete extension 'aaa' from Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(expectedEventWithExtension)),
			WantEvent:    test.CopyEventContext(e),
			Transformers: binding.TransformerFactories{DeleteExtension(extName)},
		},
		{
			Name:         "Delete extension 'aaa' from Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(expectedEventWithExtension)),
			WantEvent:    test.CopyEventContext(e),
			Transformers: binding.TransformerFactories{DeleteExtension(extName)},
		},
		{
			Name:         "Delete extension 'aaa' from Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(expectedEventWithExtension)),
			WantEvent:    test.CopyEventContext(e),
			Transformers: binding.TransformerFactories{DeleteExtension(extName)},
		},
	})
}
