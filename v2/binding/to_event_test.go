package binding_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	. "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	. "github.com/cloudevents/sdk-go/v2/test"
)

type toEventTestCase struct {
	name    string
	message binding.Message
	event   event.Event
	want    event.Event
}

func TestToEvent_success(t *testing.T) {
	EachEvent(t, Events(), func(t *testing.T, v event.Event) {
		testCases := []toEventTestCase{
			{
				name:    "From mock structured/" + TestNameOf(v),
				message: MustCreateMockStructuredMessage(t, v),
				want:    v,
			},
			{
				name:    "From mock binary/" + TestNameOf(v),
				message: MustCreateMockBinaryMessage(v),
				want:    v,
			},
			{
				name:  "From event/" + TestNameOf(v),
				event: v,
				want:  v,
			},
		}
		for _, tt := range testCases {
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
				AssertEventEquals(t, ConvertEventExtensionsToString(t, tt.want), ConvertEventExtensionsToString(t, *got))
			})
		}
	})
}

func TestToEvent_bad_spec_version_binary(t *testing.T) {
	inputEvent := FullEvent()

	inputMessage := MustCreateMockBinaryMessage(inputEvent)
	// Injecting bad spec version
	inputMessage.(*MockBinaryMessage).Metadata[spec.VS.Version(inputEvent.SpecVersion()).AttributeFromKind(spec.SpecVersion)] = "0.1.1"

	got, err := binding.ToEvent(context.Background(), inputMessage)
	require.Nil(t, got)
	require.EqualError(t, err, "unrecognized event version 0.1.1")
}

func TestToEvent_success_wrapped_event_message(t *testing.T) {
	inputEvent := FullEvent()

	cloned := inputEvent.Clone()
	inputMessage := binding.WithFinish(binding.ToMessage(&cloned), func(err error) {})

	got, err := binding.ToEvent(context.Background(), inputMessage)
	require.NoError(t, err)
	require.NotNil(t, got)
	AssertEventEquals(t, *got, inputEvent)

}

func TestToEvent_unknown(t *testing.T) {
	got, err := binding.ToEvent(context.Background(), UnknownMessage)
	require.Nil(t, got)
	require.Equal(t, binding.ErrUnknownEncoding, err)
}

func TestToEvent_wrapped_unknown(t *testing.T) {
	got, err := binding.ToEvent(context.Background(), binding.WithFinish(UnknownMessage, func(err error) {}))
	require.Nil(t, got)
	require.Equal(t, binding.ErrUnknownEncoding, err)
}

func TestToEvent_transformers_applied_once(t *testing.T) {
	EachEvent(t, Events(), func(t *testing.T, v event.Event) {
		testCases := []toEventTestCase{
			{
				name:    "From mock structured/" + TestNameOf(v),
				message: MustCreateMockStructuredMessage(t, v),
				want:    v,
			},
			{
				name:    "From mock binary/" + TestNameOf(v),
				message: MustCreateMockBinaryMessage(v),
				want:    v,
			},
			{
				name:  "From event/" + TestNameOf(v),
				event: v,
				want:  v,
			},
		}

		for _, tt := range testCases {
			t.Run("With one Transformer "+tt.name, func(t *testing.T) {
				var inputMessage binding.Message
				if tt.message != nil {
					inputMessage = tt.message
				} else {
					e := tt.event.Clone()
					inputMessage = binding.ToMessage(&e)
				}

				transformer := MockTransformer{}

				got, err := binding.ToEvent(context.Background(), inputMessage, &transformer)
				require.NoError(t, err)
				AssertEventEquals(t, ConvertEventExtensionsToString(t, tt.want), ConvertEventExtensionsToString(t, *got))

				AssertTransformerInvokedOneTime(t, &transformer)
			})
			t.Run("With two Transformers "+tt.name, func(t *testing.T) {
				var inputMessage binding.Message
				if tt.message != nil {
					inputMessage = tt.message
				} else {
					e := tt.event.Clone()
					inputMessage = binding.ToMessage(&e)
				}

				transformer1 := MockTransformer{}
				transformer2 := MockTransformer{}

				got, err := binding.ToEvent(context.Background(), inputMessage, &transformer1, &transformer2)
				require.NoError(t, err)
				AssertEventEquals(t, ConvertEventExtensionsToString(t, tt.want), ConvertEventExtensionsToString(t, *got))

				AssertTransformerInvokedOneTime(t, &transformer1)
				AssertTransformerInvokedOneTime(t, &transformer2)
			})
		}
	})
}
