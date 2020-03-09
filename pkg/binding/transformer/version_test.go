package transformer

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	test "github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/event"
	"github.com/cloudevents/sdk-go/pkg/types"

	. "github.com/cloudevents/sdk-go/pkg/bindings/test"
)

func TestVersionTranscoder(t *testing.T) {
	var testEventV03 = event.Event{
		Context: event.EventContextV03{
			Source:          types.URLRef{URL: url.URL{Path: "source"}},
			DataContentType: event.StringOfApplicationJSON(),
			ID:              "id",
			Type:            "type",
		}.AsV03(),
	}

	var testEventV1 = testEventV03
	testEventV1.Context = testEventV03.Context.AsV1()

	data := []byte("\"data\"")
	err := testEventV03.SetData(data)
	require.NoError(t, err)
	err = testEventV1.SetData(data)
	require.NoError(t, err)

	RunTransformerTests(t, context.Background(), []TransformerTestArgs{
		{
			Name:         "V03 -> V1 with Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(test.CopyEventContext(testEventV03)),
			WantEvent:    test.CopyEventContext(testEventV1),
			Transformers: binding.TransformerFactories{Version(spec.V1)},
		},
		{
			Name:         "V03 -> V1 with Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(test.CopyEventContext(testEventV03)),
			WantEvent:    test.CopyEventContext(testEventV1),
			Transformers: binding.TransformerFactories{Version(spec.V1)},
		},
		{
			Name:         "V03 -> V1 with Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(testEventV03)),
			WantEvent:    test.CopyEventContext(testEventV1),
			Transformers: binding.TransformerFactories{Version(spec.V1)},
		},
	})
}
