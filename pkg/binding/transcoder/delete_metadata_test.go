package transcoder

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
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
			name:         "Remove subject from Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(withSubjectEvent)),
			wantEvent:    noSubjectEvent,
			transformer:  binding.TransformerFactories{DeleteAttribute(spec.Subject)},
		},
		{
			name:         "Remove subject from Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(withSubjectEvent)),
			wantEvent:    noSubjectEvent,
			transformer:  binding.TransformerFactories{DeleteAttribute(spec.Subject)},
		},
		{
			name:         "Remove subject from Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(withSubjectEvent)),
			wantEvent:    noSubjectEvent,
			transformer:  binding.TransformerFactories{DeleteAttribute(spec.Subject)},
		},
		{
			name:         "Remove time from Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(withTimeEvent)),
			wantEvent:    test.CopyEventContext(withSubjectEvent),
			transformer:  binding.TransformerFactories{DeleteAttribute(spec.Time)},
		},
		{
			name:         "Remove time from Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(withTimeEvent)),
			wantEvent:    test.CopyEventContext(withSubjectEvent),
			transformer:  binding.TransformerFactories{DeleteAttribute(spec.Time)},
		},
		{
			name:         "Remove time from Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(withTimeEvent)),
			wantEvent:    test.CopyEventContext(withSubjectEvent),
			transformer:  binding.TransformerFactories{DeleteAttribute(spec.Time)},
		},
		{
			name:         "Do nothing with Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(withSubjectEvent)),
			wantEvent:    test.CopyEventContext(withSubjectEvent),
			transformer:  binding.TransformerFactories{DeleteAttribute(spec.Time)},
		},
		{
			name:         "Do nothing with Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(withSubjectEvent)),
			wantEvent:    test.CopyEventContext(withSubjectEvent),
			transformer:  binding.TransformerFactories{DeleteAttribute(spec.Time)},
		},
		{
			name:         "Do nothing with Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(withSubjectEvent)),
			wantEvent:    test.CopyEventContext(withSubjectEvent),
			transformer:  binding.TransformerFactories{DeleteAttribute(spec.Time)},
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
			name:         "No change to Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(expectedEventWithExtension)),
			wantEvent:    test.CopyEventContext(expectedEventWithExtension),
			transformer:  binding.TransformerFactories{DeleteExtension("ccc")},
		},
		{
			name:         "No change to Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(expectedEventWithExtension)),
			wantEvent:    test.CopyEventContext(expectedEventWithExtension),
			transformer:  binding.TransformerFactories{DeleteExtension("ccc")},
		},
		{
			name:         "No change to Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(expectedEventWithExtension)),
			wantEvent:    test.CopyEventContext(expectedEventWithExtension),
			transformer:  binding.TransformerFactories{DeleteExtension("ccc")},
		},
		{
			name:         "Delete extension 'aaa' from Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(expectedEventWithExtension)),
			wantEvent:    test.CopyEventContext(e),
			transformer:  binding.TransformerFactories{DeleteExtension(extName)},
		},
		{
			name:         "Delete extension 'aaa' from Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(expectedEventWithExtension)),
			wantEvent:    test.CopyEventContext(e),
			transformer:  binding.TransformerFactories{DeleteExtension(extName)},
		},
		{
			name:         "Delete extension 'aaa' from Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(expectedEventWithExtension)),
			wantEvent:    test.CopyEventContext(e),
			transformer:  binding.TransformerFactories{DeleteExtension(extName)},
		},
	})
}
