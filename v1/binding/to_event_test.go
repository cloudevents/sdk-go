package binding_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v1/binding"
	"github.com/cloudevents/sdk-go/v1/binding/test"
	"github.com/cloudevents/sdk-go/v1/cloudevents"
)

type toEventTestCase struct {
	name     string
	encoding binding.Encoding
	message  binding.Message
	want     cloudevents.Event
}

func TestToEvent(t *testing.T) {
	tests := []toEventTestCase{}

	for _, v := range test.Events() {
		tests = append(tests, []toEventTestCase{
			{
				name:     "From mock structured with payload/" + test.NameOf(v),
				encoding: binding.EncodingStructured,
				message:  test.NewMockStructuredMessage(v),
				want:     v,
			},
			{
				name:     "From mock structured without payload/" + test.NameOf(v),
				encoding: binding.EncodingStructured,
				message:  test.NewMockStructuredMessage(v),
				want:     v,
			},
			{
				name:     "From mock binary with payload/" + test.NameOf(v),
				encoding: binding.EncodingBinary,
				message:  test.NewMockBinaryMessage(v),
				want:     v,
			},
			{
				name:     "From mock binary without payload/" + test.NameOf(v),
				encoding: binding.EncodingBinary,
				message:  test.NewMockBinaryMessage(v),
				want:     v,
			},
			{
				name:     "From event with payload/" + test.NameOf(v),
				encoding: binding.EncodingEvent,
				message:  binding.EventMessage(v),
				want:     v,
			},
			{
				name:     "From event without payload/" + test.NameOf(v),
				encoding: binding.EncodingEvent,
				message:  binding.EventMessage(v),
				want:     v,
			},
		}...)
	}
	for _, tt := range tests {
		tt := tt // Don't use range variable in Run() scope
		t.Run(tt.name, func(t *testing.T) {
			got, encoding, err := binding.ToEvent(context.Background(), tt.message)
			require.NoError(t, err)
			require.Equal(t, tt.encoding, encoding)
			test.AssertEventEquals(t, test.ExToStr(t, tt.want), test.ExToStr(t, got))
		})
	}
}
