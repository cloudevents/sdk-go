package transcoder

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
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

	RunTranscoderTests(t, context.Background(), []TranscoderTestArgs{
		{
			name:         "No change to id to Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(e)),
			wantEvent:    test.CopyEventContext(e),
			transformer:  binding.TransformerFactories{AddAttribute(spec.ID, "new-id")},
		},
		{
			name:         "No change to id to Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(e)),
			wantEvent:    test.CopyEventContext(e),
			transformer:  binding.TransformerFactories{AddAttribute(spec.ID, "new-id")},
		},
		{
			name:         "No change to id to Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(e)),
			wantEvent:    test.CopyEventContext(e),
			transformer:  binding.TransformerFactories{AddAttribute(spec.ID, "new-id")},
		},
		{
			name:         "Add subject to Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(e)),
			wantEvent:    expectedEventWithSubject,
			transformer:  binding.TransformerFactories{AddAttribute(spec.Subject, subject)},
		},
		{
			name:         "Add subject to Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(e)),
			wantEvent:    expectedEventWithSubject,
			transformer:  binding.TransformerFactories{AddAttribute(spec.Subject, subject)},
		},
		{
			name:         "Add subject to Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(e)),
			wantEvent:    expectedEventWithSubject,
			transformer:  binding.TransformerFactories{AddAttribute(spec.Subject, subject)},
		},
		{
			name:         "Add time to Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(e)),
			wantEvent:    expectedEventWithTime,
			transformer:  binding.TransformerFactories{AddAttribute(spec.Time, timestamp)},
		},
		{
			name:         "Add time to Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(e)),
			wantEvent:    expectedEventWithTime,
			transformer:  binding.TransformerFactories{AddAttribute(spec.Time, timestamp)},
		},
		{
			name:         "Add time to Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(e)),
			wantEvent:    expectedEventWithTime,
			transformer:  binding.TransformerFactories{AddAttribute(spec.Time, timestamp)},
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

	RunTranscoderTests(t, context.Background(), []TranscoderTestArgs{
		{
			name:         "No change to extension 'aaa' to Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(expectedEventWithExtension)),
			wantEvent:    test.CopyEventContext(expectedEventWithExtension),
			transformer:  binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			name:         "No change to extension 'aaa' to Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(expectedEventWithExtension)),
			wantEvent:    test.CopyEventContext(expectedEventWithExtension),
			transformer:  binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			name:         "No change to extension 'aaa' to Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(expectedEventWithExtension)),
			wantEvent:    test.CopyEventContext(expectedEventWithExtension),
			transformer:  binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			name:         "Add extension 'aaa' to Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(e)),
			wantEvent:    test.CopyEventContext(expectedEventWithExtension),
			transformer:  binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			name:         "Add extension 'aaa' to Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(e)),
			wantEvent:    test.CopyEventContext(expectedEventWithExtension),
			transformer:  binding.TransformerFactories{AddExtension(extName, extValue)},
		},
		{
			name:         "Add extension 'aaa' to Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(e)),
			wantEvent:    test.CopyEventContext(expectedEventWithExtension),
			transformer:  binding.TransformerFactories{AddExtension(extName, extValue)},
		},
	})
}
