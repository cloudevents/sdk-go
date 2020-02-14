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

	subjectUpdateFunc := func(v interface{}) interface{} {
		return strings.ToUpper(v.(string))
	}
	updatedSubjectEvent := copyEventContext(withSubjectEvent)
	require.NoError(t, updatedSubjectEvent.Context.SetSubject(subjectUpdateFunc("abc").(string)))

	location, err := time.LoadLocation("UTC")
	require.NoError(t, err)
	timestamp := time.Now().In(location)
	withTimeEvent := copyEventContext(withSubjectEvent)
	require.NoError(t, withTimeEvent.Context.SetTime(timestamp))

	timeUpdateFunc := func(v interface{}) interface{} {
		return v.(time.Time).Add(3 * time.Hour)
	}
	updatedTimeEvent := copyEventContext(withTimeEvent)
	require.NoError(t, updatedTimeEvent.Context.SetTime(timeUpdateFunc(timestamp).(time.Time)))

	RunTranscoderTests(t, context.Background(), []TranscoderTestArgs{
		{
			name:         "Update subject in Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(copyEventContext(withSubjectEvent)),
			wantEvent:    copyEventContext(updatedSubjectEvent),
			transformer:  UpdateAttribute(spec.Subject, subjectUpdateFunc),
		},
		{
			name:         "Update subject in Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(copyEventContext(withSubjectEvent)),
			wantEvent:    copyEventContext(updatedSubjectEvent),
			transformer:  UpdateAttribute(spec.Subject, subjectUpdateFunc),
		},
		{
			name:         "Update subject in Event message",
			inputMessage: binding.EventMessage(copyEventContext(withSubjectEvent)),
			wantEvent:    copyEventContext(updatedSubjectEvent),
			transformer:  UpdateAttribute(spec.Subject, subjectUpdateFunc),
		},
		{
			name:         "Update time in Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(copyEventContext(withTimeEvent)),
			wantEvent:    copyEventContext(updatedTimeEvent),
			transformer:  UpdateAttribute(spec.Time, timeUpdateFunc),
		},
		{
			name:         "Update time in Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(copyEventContext(withTimeEvent)),
			wantEvent:    copyEventContext(updatedTimeEvent),
			transformer:  UpdateAttribute(spec.Time, timeUpdateFunc),
		},
		{
			name:         "Update time in Event message",
			inputMessage: binding.EventMessage(copyEventContext(withTimeEvent)),
			wantEvent:    copyEventContext(updatedTimeEvent),
			transformer:  UpdateAttribute(spec.Time, timeUpdateFunc),
		},
		{
			name:         "Do nothing with Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(copyEventContext(withSubjectEvent)),
			wantEvent:    copyEventContext(withSubjectEvent),
			transformer: UpdateAttribute(spec.DataContentType, func(i interface{}) interface{} {
				return "text/plain"
			}),
		},
		{
			name:         "Do nothing with Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(copyEventContext(withSubjectEvent)),
			wantEvent:    copyEventContext(withSubjectEvent),
			transformer: UpdateAttribute(spec.DataContentType, func(i interface{}) interface{} {
				return "text/plain"
			}),
		},
		{
			name:         "Do nothing with Event message",
			inputMessage: binding.EventMessage(copyEventContext(withSubjectEvent)),
			wantEvent:    copyEventContext(withSubjectEvent),
			transformer: UpdateAttribute(spec.DataContentType, func(i interface{}) interface{} {
				return "text/plain"
			}),
		},
	})
}

func TestUpdateExtension(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()
	require.NoError(t, e.Context.SetExtension("aaa", "bbb"))

	updateFunc := func(v interface{}) interface{} {
		return strings.ToUpper(v.(string))
	}
	updatedExtensionEvent := copyEventContext(e)
	require.NoError(t, updatedExtensionEvent.Context.SetExtension("aaa", updateFunc("bbb").(string)))

	RunTranscoderTests(t, context.Background(), []TranscoderTestArgs{
		{
			name:         "No change in Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(copyEventContext(e)),
			wantEvent:    copyEventContext(e),
			transformer:  UpdateExtension("ccc", updateFunc),
		},
		{
			name:         "No change in Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(copyEventContext(e)),
			wantEvent:    copyEventContext(e),
			transformer:  UpdateExtension("ccc", updateFunc),
		},
		{
			name:         "No change in Event message",
			inputMessage: binding.EventMessage(copyEventContext(e)),
			wantEvent:    copyEventContext(e),
			transformer:  UpdateExtension("ccc", updateFunc),
		},
		{
			name:         "Update extension 'aaa' in Mock Structured message",
			inputMessage: test.NewMockStructuredMessage(copyEventContext(e)),
			wantEvent:    copyEventContext(updatedExtensionEvent),
			transformer:  UpdateExtension("aaa", updateFunc),
		},
		{
			name:         "Update extension 'aaa' in Mock Binary message",
			inputMessage: test.NewMockBinaryMessage(copyEventContext(e)),
			wantEvent:    copyEventContext(updatedExtensionEvent),
			transformer:  UpdateExtension("aaa", updateFunc),
		},
		{
			name:         "Update extension 'aaa' in Event message",
			inputMessage: binding.EventMessage(copyEventContext(e)),
			wantEvent:    copyEventContext(updatedExtensionEvent),
			transformer:  UpdateExtension("aaa", updateFunc),
		},
	})
}
