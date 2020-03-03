package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/event"
)

type TranscoderTestArgs struct {
	Name         string
	InputMessage binding.Message
	WantEvent    event.Event
	AssertFunc   func(t *testing.T, event event.Event)
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

			var e event.Event
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
			if tt.AssertFunc != nil {
				tt.AssertFunc(t, e)
			} else {
				AssertEventEquals(t, tt.WantEvent, e)
			}
		})
	}
}
