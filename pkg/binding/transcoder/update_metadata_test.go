package transcoder

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
)

func TestUpdateAttribute(t *testing.T) {
	withSubjectEvent := test.MinEvent()
	withSubjectEvent.Context = withSubjectEvent.Context.AsV1()
	require.NoError(t, withSubjectEvent.Context.SetSubject("abc"))

	subjectUpdateFunc := func(v interface{}) (interface{}, error) {
		return strings.ToUpper(v.(string)), nil
	}
	updatedSubjectEvent := test.CopyEventContext(withSubjectEvent)
	require.NoError(t, updatedSubjectEvent.Context.SetSubject(strings.ToUpper("abc")))

	location, err := time.LoadLocation("UTC")
	require.NoError(t, err)
	timestamp := time.Now().In(location)
	withTimeEvent := test.CopyEventContext(withSubjectEvent)
	require.NoError(t, withTimeEvent.Context.SetTime(timestamp))

	timeUpdateFunc := func(v interface{}) (interface{}, error) {
		return v.(time.Time).Add(3 * time.Hour), nil
	}
	updatedTimeEvent := test.CopyEventContext(withTimeEvent)
	require.NoError(t, updatedTimeEvent.Context.SetTime(timestamp.Add(3*time.Hour)))

	test.RunTranscoderTests(t, context.Background(), []test.TranscoderTestArgs{
		{
			name:         "Update subject in Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(withSubjectEvent)),
			wantEvent:    test.CopyEventContext(updatedSubjectEvent),
			transformer:  binding.TransformerFactories{UpdateAttribute(spec.Subject, subjectUpdateFunc)},
		},
		{
			name:         "Update subject in Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(withSubjectEvent)),
			wantEvent:    test.CopyEventContext(updatedSubjectEvent),
			transformer:  binding.TransformerFactories{UpdateAttribute(spec.Subject, subjectUpdateFunc)},
		},
		{
			name:         "Update subject in Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(withSubjectEvent)),
			wantEvent:    test.CopyEventContext(updatedSubjectEvent),
			transformer:  binding.TransformerFactories{UpdateAttribute(spec.Subject, subjectUpdateFunc)},
		},
		{
			name:         "Update time in Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(withTimeEvent)),
			wantEvent:    test.CopyEventContext(updatedTimeEvent),
			transformer:  binding.TransformerFactories{UpdateAttribute(spec.Time, timeUpdateFunc)},
		},
		{
			name:         "Update time in Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(withTimeEvent)),
			wantEvent:    test.CopyEventContext(updatedTimeEvent),
			transformer:  binding.TransformerFactories{UpdateAttribute(spec.Time, timeUpdateFunc)},
		},
		{
			name:         "Update time in Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(withTimeEvent)),
			wantEvent:    test.CopyEventContext(updatedTimeEvent),
			transformer:  binding.TransformerFactories{UpdateAttribute(spec.Time, timeUpdateFunc)},
		},
		{
			name:         "Do nothing with Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(withSubjectEvent)),
			wantEvent:    test.CopyEventContext(withSubjectEvent),
			transformer: binding.TransformerFactories{UpdateAttribute(spec.DataContentType, func(i interface{}) (interface{}, error) {
				return "text/plain", nil
			})},
		},
		{
			name:         "Do nothing with Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(withSubjectEvent)),
			wantEvent:    test.CopyEventContext(withSubjectEvent),
			transformer: binding.TransformerFactories{UpdateAttribute(spec.DataContentType, func(i interface{}) (interface{}, error) {
				return "text/plain", nil
			})},
		},
		{
			name:         "Do nothing with Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(withSubjectEvent)),
			wantEvent:    test.CopyEventContext(withSubjectEvent),
			transformer: binding.TransformerFactories{UpdateAttribute(spec.DataContentType, func(i interface{}) (interface{}, error) {
				return "text/plain", nil
			})},
		},
	})
}

func TestUpdateExtension(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()
	require.NoError(t, e.Context.SetExtension("aaa", "bbb"))

	updateFunc := func(v interface{}) (interface{}, error) {
		return strings.ToUpper(v.(string)), nil
	}
	updatedExtensionEvent := test.CopyEventContext(e)
	require.NoError(t, updatedExtensionEvent.Context.SetExtension("aaa", strings.ToUpper("bbb")))

	test.RunTranscoderTests(t, context.Background(), []test.TranscoderTestArgs{
		{
			name:         "No change in Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(e)),
			wantEvent:    test.CopyEventContext(e),
			transformer:  binding.TransformerFactories{UpdateExtension("ccc", updateFunc)},
		},
		{
			name:         "No change in Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(e)),
			wantEvent:    test.CopyEventContext(e),
			transformer:  binding.TransformerFactories{UpdateExtension("ccc", updateFunc)},
		},
		{
			name:         "No change in Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(e)),
			wantEvent:    test.CopyEventContext(e),
			transformer:  binding.TransformerFactories{UpdateExtension("ccc", updateFunc)},
		},
		{
			name:         "Update extension 'aaa' in Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(test.CopyEventContext(e)),
			wantEvent:    test.CopyEventContext(updatedExtensionEvent),
			transformer:  binding.TransformerFactories{UpdateExtension("aaa", updateFunc)},
		},
		{
			name:         "Update extension 'aaa' in Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(test.CopyEventContext(e)),
			wantEvent:    test.CopyEventContext(updatedExtensionEvent),
			transformer:  binding.TransformerFactories{UpdateExtension("aaa", updateFunc)},
		},
		{
			name:         "Update extension 'aaa' in Event message",
			inputMessage: binding.EventMessage(test.CopyEventContext(e)),
			wantEvent:    test.CopyEventContext(updatedExtensionEvent),
			transformer:  binding.TransformerFactories{UpdateExtension("aaa", updateFunc)},
		},
	})
}
