package transcoder

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	cloudevents "github.com/cloudevents/sdk-go/v1"
	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/spec"
	"github.com/cloudevents/sdk-go/v1/binding/test"
	"github.com/cloudevents/sdk-go/v1/cloudevents/types"
)

func TestVersionTranscoder(t *testing.T) {
	var testEventV02 = cloudevents.Event{
		Context: cloudevents.EventContextV02{
			Source:      types.URLRef{URL: url.URL{Path: "source"}},
			ContentType: cloudevents.StringOfApplicationJSON(),
			ID:          "id",
			Type:        "type",
		}.AsV02(),
	}

	var testEventV1 = testEventV02
	testEventV1.Context = testEventV02.Context.AsV1()

	data := []byte("\"data\"")
	err := testEventV02.SetData(data)
	require.NoError(t, err)
	err = testEventV1.SetData(data)
	require.NoError(t, err)

	test.RunTranscoderTests(t, context.Background(), []test.TranscoderTestArgs{
		{
			Name:         "V02 -> V1 with Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(test.CopyEventContext(testEventV02)),
			WantEvent:    test.CopyEventContext(testEventV1),
			Transformers: binding.TransformerFactories{Version(spec.V1)},
		},
		{
			Name:         "V02 -> V1 with Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(test.CopyEventContext(testEventV02)),
			WantEvent:    test.CopyEventContext(testEventV1),
			Transformers: binding.TransformerFactories{Version(spec.V1)},
		},
		{
			Name:         "V02 -> V1 with Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(testEventV02)),
			WantEvent:    test.CopyEventContext(testEventV1),
			Transformers: binding.TransformerFactories{Version(spec.V1)},
		},
	})
}
