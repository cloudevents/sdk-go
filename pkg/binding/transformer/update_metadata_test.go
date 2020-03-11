package transformer

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	. "github.com/cloudevents/sdk-go/pkg/binding/test"
)

func TestUpdateAttribute(t *testing.T) {
	withSubjectEvent := MinEvent()
	withSubjectEvent.Context = withSubjectEvent.Context.AsV1()
	require.NoError(t, withSubjectEvent.Context.SetSubject("abc"))

	subjectUpdateFunc := func(v interface{}) (interface{}, error) {
		return strings.ToUpper(v.(string)), nil
	}
	updatedSubjectEvent := CopyEventContext(withSubjectEvent)
	require.NoError(t, updatedSubjectEvent.Context.SetSubject(strings.ToUpper("abc")))

	location, err := time.LoadLocation("UTC")
	require.NoError(t, err)
	timestamp := time.Now().In(location)
	withTimeEvent := CopyEventContext(withSubjectEvent)
	require.NoError(t, withTimeEvent.Context.SetTime(timestamp))

	timeUpdateFunc := func(v interface{}) (interface{}, error) {
		return v.(time.Time).Add(3 * time.Hour), nil
	}
	updatedTimeEvent := CopyEventContext(withTimeEvent)
	require.NoError(t, updatedTimeEvent.Context.SetTime(timestamp.Add(3*time.Hour)))

	RunTransformerTests(t, context.Background(), []TransformerTestArgs{
		{
			Name:         "Update subject in Mock Structured message",
			InputMessage: MustCreateMockStructuredMessage(CopyEventContext(withSubjectEvent)),
			WantEvent:    CopyEventContext(updatedSubjectEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.Subject, subjectUpdateFunc)},
		},
		{
			Name:         "Update subject in Mock Binary message",
			InputMessage: MustCreateMockBinaryMessage(CopyEventContext(withSubjectEvent)),
			WantEvent:    CopyEventContext(updatedSubjectEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.Subject, subjectUpdateFunc)},
		},
		{
			Name:         "Update subject in Event message",
			InputEvent:   withSubjectEvent,
			WantEvent:    CopyEventContext(updatedSubjectEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.Subject, subjectUpdateFunc)},
		},
		{
			Name:         "Update time in Mock Structured message",
			InputMessage: MustCreateMockStructuredMessage(CopyEventContext(withTimeEvent)),
			WantEvent:    CopyEventContext(updatedTimeEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.Time, timeUpdateFunc)},
		},
		{
			Name:         "Update time in Mock Binary message",
			InputMessage: MustCreateMockBinaryMessage(CopyEventContext(withTimeEvent)),
			WantEvent:    CopyEventContext(updatedTimeEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.Time, timeUpdateFunc)},
		},
		{
			Name:         "Update time in Event message",
			InputEvent:   withTimeEvent,
			WantEvent:    CopyEventContext(updatedTimeEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.Time, timeUpdateFunc)},
		},
		{
			Name:         "Do nothing with Mock Structured message",
			InputMessage: MustCreateMockStructuredMessage(CopyEventContext(withSubjectEvent)),
			WantEvent:    CopyEventContext(withSubjectEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.DataContentType, func(i interface{}) (interface{}, error) {
				return "text/plain", nil
			})},
		},
		{
			Name:         "Do nothing with Mock Binary message",
			InputMessage: MustCreateMockBinaryMessage(CopyEventContext(withSubjectEvent)),
			WantEvent:    CopyEventContext(withSubjectEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.DataContentType, func(i interface{}) (interface{}, error) {
				return "text/plain", nil
			})},
		},
		{
			Name:       "Do nothing with Event message",
			InputEvent: withSubjectEvent,
			WantEvent:  CopyEventContext(withSubjectEvent),
			Transformers: binding.TransformerFactories{UpdateAttribute(spec.DataContentType, func(i interface{}) (interface{}, error) {
				return "text/plain", nil
			})},
		},
	})
}

func TestUpdateExtension(t *testing.T) {
	e := MinEvent()
	e.Context = e.Context.AsV1()
	require.NoError(t, e.Context.SetExtension("aaa", "bbb"))

	updateFunc := func(v interface{}) (interface{}, error) {
		return strings.ToUpper(v.(string)), nil
	}
	updatedExtensionEvent := CopyEventContext(e)
	require.NoError(t, updatedExtensionEvent.Context.SetExtension("aaa", strings.ToUpper("bbb")))

	RunTransformerTests(t, context.Background(), []TransformerTestArgs{
		{
			Name:         "No change in Mock Structured message",
			InputMessage: MustCreateMockStructuredMessage(CopyEventContext(e)),
			WantEvent:    CopyEventContext(e),
			Transformers: binding.TransformerFactories{UpdateExtension("ccc", updateFunc)},
		},
		{
			Name:         "No change in Mock Binary message",
			InputMessage: MustCreateMockBinaryMessage(CopyEventContext(e)),
			WantEvent:    CopyEventContext(e),
			Transformers: binding.TransformerFactories{UpdateExtension("ccc", updateFunc)},
		},
		{
			Name:         "No change in Event message",
			InputEvent:   e,
			WantEvent:    CopyEventContext(e),
			Transformers: binding.TransformerFactories{UpdateExtension("ccc", updateFunc)},
		},
		{
			Name:         "Update extension 'aaa' in Mock Structured message",
			InputEvent:   e,
			WantEvent:    CopyEventContext(updatedExtensionEvent),
			Transformers: binding.TransformerFactories{UpdateExtension("aaa", updateFunc)},
		},
		{
			Name:         "Update extension 'aaa' in Mock Binary message",
			InputMessage: MustCreateMockBinaryMessage(CopyEventContext(e)),
			WantEvent:    CopyEventContext(updatedExtensionEvent),
			Transformers: binding.TransformerFactories{UpdateExtension("aaa", updateFunc)},
		},
		{
			Name:         "Update extension 'aaa' in Event message",
			InputEvent:   e,
			WantEvent:    CopyEventContext(updatedExtensionEvent),
			Transformers: binding.TransformerFactories{UpdateExtension("aaa", updateFunc)},
		},
	})
}
