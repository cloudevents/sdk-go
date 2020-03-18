package transcoder

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/spec"
	"github.com/cloudevents/sdk-go/v1/binding/test"
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
			Name:         "Update subject in Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(withSubjectEvent)),
			WantEvent:    test.CopyEventContext(updatedSubjectEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.Subject, subjectUpdateFunc)},
		},
		{
			Name:         "Update subject in Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(withSubjectEvent)),
			WantEvent:    test.CopyEventContext(updatedSubjectEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.Subject, subjectUpdateFunc)},
		},
		{
			Name:         "Update subject in Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(withSubjectEvent)),
			WantEvent:    test.CopyEventContext(updatedSubjectEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.Subject, subjectUpdateFunc)},
		},
		{
			Name:         "Update time in Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(withTimeEvent)),
			WantEvent:    test.CopyEventContext(updatedTimeEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.Time, timeUpdateFunc)},
		},
		{
			Name:         "Update time in Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(withTimeEvent)),
			WantEvent:    test.CopyEventContext(updatedTimeEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.Time, timeUpdateFunc)},
		},
		{
			Name:         "Update time in Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(withTimeEvent)),
			WantEvent:    test.CopyEventContext(updatedTimeEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.Time, timeUpdateFunc)},
		},
		{
			Name:         "Do nothing with Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(withSubjectEvent)),
			WantEvent:    test.CopyEventContext(withSubjectEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.DataContentType, func(i interface{}) (interface{}, error) {
				return "text/plain", nil
			})},
		},
		{
			Name:         "Do nothing with Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(withSubjectEvent)),
			WantEvent:    test.CopyEventContext(withSubjectEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.DataContentType, func(i interface{}) (interface{}, error) {
				return "text/plain", nil
			})},
		},
		{
			Name:         "Do nothing with Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(withSubjectEvent)),
			WantEvent:    test.CopyEventContext(withSubjectEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.DataContentType, func(i interface{}) (interface{}, error) {
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
			Name:         "No change in Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(e)),
			WantEvent:    test.CopyEventContext(e),
			Transformers: binding.TransformerFactories{UpdateExtension("ccc", updateFunc)},
		},
		{
			Name:         "No change in Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(e)),
			WantEvent:    test.CopyEventContext(e),
			Transformers: binding.TransformerFactories{UpdateExtension("ccc", updateFunc)},
		},
		{
			Name:         "No change in Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(e)),
			WantEvent:    test.CopyEventContext(e),
			Transformers: binding.TransformerFactories{UpdateExtension("ccc", updateFunc)},
		},
		{
			Name:         "Update extension 'aaa' in Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(e)),
			WantEvent:    test.CopyEventContext(updatedExtensionEvent),
			Transformers: binding.TransformerFactories{UpdateExtension("aaa", updateFunc)},
		},
		{
			Name:         "Update extension 'aaa' in Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(e)),
			WantEvent:    test.CopyEventContext(updatedExtensionEvent),
			Transformers: binding.TransformerFactories{UpdateExtension("aaa", updateFunc)},
		},
		{
			Name:         "Update extension 'aaa' in Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(e)),
			WantEvent:    test.CopyEventContext(updatedExtensionEvent),
			Transformers: binding.TransformerFactories{UpdateExtension("aaa", updateFunc)},
		},
	})
}
