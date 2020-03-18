package binding_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
)

type toEventTestCase struct {
	name    string
	message binding.Message
	event   event.Event
	want    event.Event
}

func TestToEvent_success(t *testing.T) {
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
				name:  "From event with payload/" + test.NameOf(v),
				event: v,
				want:  v,
			},
			{
				name:  "From event without payload/" + test.NameOf(v),
				event: v,
				want:  v,
			},
		}...)
	}
	for _, tt := range tests {
		tt := tt // Don't use range variable in Run() scope
		t.Run(tt.name, func(t *testing.T) {
			var inputMessage binding.Message
			if tt.message != nil {
				inputMessage = tt.message
			} else {
				e := tt.event.Clone()
				inputMessage = binding.ToMessage(&e)
			}
			got, err := binding.ToEvent(context.Background(), inputMessage)
			require.NoError(t, err)
			require.NotNil(t, got)
			test.AssertEventEquals(t, test.ExToStr(t, tt.want), test.ExToStr(t, *got))
		})
	}
}

func TestToEvent_success_wrapped_event_message(t *testing.T) {
	inputEvent := test.FullEvent()

	cloned := inputEvent.Clone()
	inputMessage := binding.WithFinish(binding.ToMessage(&cloned), func(err error) {})

	got, err := binding.ToEvent(context.Background(), inputMessage)
	require.NoError(t, err)
	require.NotNil(t, got)
	test.AssertEventEquals(t, *got, inputEvent)

}

func TestToEvent_unknown(t *testing.T) {
	got, err := binding.ToEvent(context.Background(), test.UnknownMessage)
	require.Nil(t, got)
	require.Equal(t, binding.ErrUnknownEncoding, err)
}

func TestToEvent_wrapped_unknown(t *testing.T) {
	got, err := binding.ToEvent(context.Background(), binding.WithFinish(test.UnknownMessage, func(err error) {}))
	require.Nil(t, got)
	require.Equal(t, binding.ErrUnknownEncoding, err)
}
