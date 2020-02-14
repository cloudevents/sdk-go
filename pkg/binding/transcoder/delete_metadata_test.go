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

	withTimeEvent := copyEventContext(withSubjectEvent)
	require.NoError(t, withTimeEvent.Context.SetTime(time.Now()))

	noSubjectEvent := copyEventContext(withSubjectEvent)
	require.NoError(t, noSubjectEvent.Context.SetSubject(""))

	RunTranscoderTests(t, context.Background(), []TranscoderTestArgs{
		{
			name:         "Remove subject from Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(copyEventContext(withSubjectEvent)),
			wantEvent:    noSubjectEvent,
			transformer:  DeleteAttribute(spec.Subject),
		},
		{
			name:         "Remove subject from Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(copyEventContext(withSubjectEvent)),
			wantEvent:    noSubjectEvent,
			transformer:  DeleteAttribute(spec.Subject),
		},
		{
			name:         "Remove subject from Event message",
			inputMessage: binding.EventMessage(copyEventContext(withSubjectEvent)),
			wantEvent:    noSubjectEvent,
			transformer:  DeleteAttribute(spec.Subject),
		},
		{
			name:         "Remove time from Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(copyEventContext(withTimeEvent)),
			wantEvent:    copyEventContext(withSubjectEvent),
			transformer:  DeleteAttribute(spec.Time),
		},
		{
			name:         "Remove time from Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(copyEventContext(withTimeEvent)),
			wantEvent:    copyEventContext(withSubjectEvent),
			transformer:  DeleteAttribute(spec.Time),
		},
		{
			name:         "Remove time from Event message",
			inputMessage: binding.EventMessage(copyEventContext(withTimeEvent)),
			wantEvent:    copyEventContext(withSubjectEvent),
			transformer:  DeleteAttribute(spec.Time),
		},
		{
			name:         "Do nothing with Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(copyEventContext(withSubjectEvent)),
			wantEvent:    copyEventContext(withSubjectEvent),
			transformer:  DeleteAttribute(spec.Time),
		},
		{
			name:         "Do nothing with Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(copyEventContext(withSubjectEvent)),
			wantEvent:    copyEventContext(withSubjectEvent),
			transformer:  DeleteAttribute(spec.Time),
		},
		{
			name:         "Do nothing with Event message",
			inputMessage: binding.EventMessage(copyEventContext(withSubjectEvent)),
			wantEvent:    copyEventContext(withSubjectEvent),
			transformer:  DeleteAttribute(spec.Time),
		},
	})
}

func TestDeleteExtension(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	extName := "aaa"
	extValue := "bbb"
	expectedEventWithExtension := copyEventContext(e)
	require.NoError(t, expectedEventWithExtension.Context.SetExtension(extName, extValue))

	RunTranscoderTests(t, context.Background(), []TranscoderTestArgs{
		{
			name:         "No change to Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(copyEventContext(expectedEventWithExtension)),
			wantEvent:    copyEventContext(expectedEventWithExtension),
			transformer:  DeleteExtension("ccc"),
		},
		{
			name:         "No change to Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(copyEventContext(expectedEventWithExtension)),
			wantEvent:    copyEventContext(expectedEventWithExtension),
			transformer:  DeleteExtension("ccc"),
		},
		{
			name:         "No change to Event message",
			inputMessage: binding.EventMessage(copyEventContext(expectedEventWithExtension)),
			wantEvent:    copyEventContext(expectedEventWithExtension),
			transformer:  DeleteExtension("ccc"),
		},
		{
			name:         "Delete extension 'aaa' from Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(copyEventContext(expectedEventWithExtension)),
			wantEvent:    copyEventContext(e),
			transformer:  DeleteExtension(extName),
		},
		{
			name:         "Delete extension 'aaa' from Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(copyEventContext(expectedEventWithExtension)),
			wantEvent:    copyEventContext(e),
			transformer:  DeleteExtension(extName),
		},
		{
			name:         "Delete extension 'aaa' from Event message",
			inputMessage: binding.EventMessage(copyEventContext(expectedEventWithExtension)),
			wantEvent:    copyEventContext(e),
			transformer:  DeleteExtension(extName),
		},
	})
}
