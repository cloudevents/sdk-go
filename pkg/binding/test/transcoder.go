package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/buffering"
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
			copied, err := buffering.CopyMessage(ctx, tt.InputMessage, tt.Transformers...)
			require.NoError(t, err)
			e, _, err := binding.ToEvent(ctx, copied)
			require.NoError(t, err)
			AssertEventEquals(t, tt.WantEvent, e)
		})
	}
}
