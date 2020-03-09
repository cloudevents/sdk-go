package binding_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/event"
)

type toEventTestCase struct {
	name    string
	message binding.Message
	want    event.Event
}

func TestToEvent(t *testing.T) {
	tests := []toEventTestCase{}

	for _, v := range test.Events() {
		tests = append(tests, []toEventTestCase{
			{
				name:    "From mock structured with payload/" + test.NameOf(v),
				message: test.MustCreateMockStructuredMessage(v),
				want:    v,
			},
			{
				name:    "From mock structured without payload/" + test.NameOf(v),
				message: test.MustCreateMockStructuredMessage(v),
				want:    v,
			},
			{
				name:    "From mock binary with payload/" + test.NameOf(v),
				message: test.MustCreateMockBinaryMessage(v),
				want:    v,
			},
			{
				name:    "From mock binary without payload/" + test.NameOf(v),
				message: test.MustCreateMockBinaryMessage(v),
				want:    v,
			},
			{
				name:    "From event with payload/" + test.NameOf(v),
				message: binding.EventMessage(v),
				want:    v,
			},
			{
				name:    "From event without payload/" + test.NameOf(v),
				message: binding.EventMessage(v),
				want:    v,
			},
		}...)
	}
	for _, tt := range tests {
		tt := tt // Don't use range variable in Run() scope
		t.Run(tt.name, func(t *testing.T) {
			got, err := binding.ToEvent(context.Background(), tt.message, nil)
			require.NoError(t, err)
			test.AssertEventEquals(t, test.ExToStr(t, tt.want), test.ExToStr(t, got))
		})
	}
}
