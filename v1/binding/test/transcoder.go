package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cloudevents "github.com/cloudevents/sdk-go/v1"
	"github.com/cloudevents/sdk-go/v1/binding"
)

type TranscoderTestArgs struct {
	Name         string
	InputMessage binding.Message
	WantEvent    cloudevents.Event
	Transformers []binding.TransformerFactory
}

func RunTranscoderTests(t *testing.T, ctx context.Context, tests []TranscoderTestArgs) {
	for _, tt := range tests {
		tt := tt // Don't use range variable inside scope
		t.Run(tt.Name, func(t *testing.T) {

			mockStructured := MockStructuredMessage{}
			mockBinary := MockBinaryMessage{}

			enc, err := binding.Encode(ctx, tt.InputMessage, &mockStructured, &mockBinary, tt.Transformers)
			require.NoError(t, err)

			var e cloudevents.Event
			if enc == binding.EncodingStructured {
				e, _, err = binding.ToEvent(ctx, &mockStructured)
				require.NoError(t, err)
			} else if enc == binding.EncodingBinary {
				e, _, err = binding.ToEvent(ctx, &mockBinary)
				require.NoError(t, err)
			} else {
				t.Fatalf("Unexpected encoding %v", enc)
			}
			require.NoError(t, err)
			AssertEventEquals(t, tt.WantEvent, e)
		})
	}
}
