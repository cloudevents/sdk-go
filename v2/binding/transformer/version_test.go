package transformer

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

func TestVersionTranscoder(t *testing.T) {
	var testEventV03 = event.Event{
		Context: event.EventContextV03{
			Source: types.URIRef{URL: url.URL{Path: "source"}},
			ID:     "id",
			Type:   "type",
		}.AsV03(),
	}

	var testEventV1 = testEventV03
	testEventV1.Context = testEventV03.Context.AsV1()

	data := []byte("\"data\"")
	err := testEventV03.SetData(event.ApplicationJSON, data)
	require.NoError(t, err)
	err = testEventV1.SetData(event.ApplicationJSON, data)
	require.NoError(t, err)

	test.RunTransformerTests(t, context.Background(), []test.TransformerTestArgs{
		{
			Name:         "V03 -> V1 with Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(t, testEventV03),
			WantEvent:    testEventV1,
			Transformers: binding.Transformers{Version(spec.V1)},
		},
		{
			Name:         "V03 -> V1 with Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(testEventV03),
			WantEvent:    testEventV1,
			Transformers: binding.Transformers{Version(spec.V1)},
		},
		{
			Name:         "V03 -> V1 with Event message",
			InputEvent:   testEventV03,
			WantEvent:    testEventV1,
			Transformers: binding.Transformers{Version(spec.V1)},
		},
	})
}
