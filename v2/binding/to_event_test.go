package binding_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
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
	test.EachEvent(t, test.Events(), func(t *testing.T, v event.Event) {
		testCases := []toEventTestCase{
			{
				name:    "From mock structured/" + test.NameOf(v),
				message: test.MustCreateMockStructuredMessage(v),
				want:    v,
			},
			{
				name:    "From mock binary/" + test.NameOf(v),
				message: test.MustCreateMockBinaryMessage(v),
				want:    v,
			},
			{
				name:  "From event/" + test.NameOf(v),
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
				test.AssertEventEquals(t, test.ExToStr(t, tt.want), test.ExToStr(t, *got))
			})
		}
	})
}

func TestToEvent_bad_spec_version_binary(t *testing.T) {
	inputEvent := test.FullEvent()

	inputMessage := test.MustCreateMockBinaryMessage(inputEvent)
	// Injecting bad spec version
	inputMessage.(*test.MockBinaryMessage).Metadata[spec.VS.Version(inputEvent.SpecVersion()).AttributeFromKind(spec.SpecVersion)] = "0.1.1"

	got, err := binding.ToEvent(context.Background(), inputMessage)
	require.Nil(t, got)
	require.EqualError(t, err, "unrecognized event version 0.1.1")
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

func TestToEvent_transformers_applied_once(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, v event.Event) {
		testCases := []toEventTestCase{
			{
				name:    "From mock structured/" + test.NameOf(v),
				message: test.MustCreateMockStructuredMessage(v),
				want:    v,
			},
			{
				name:    "From mock binary/" + test.NameOf(v),
				message: test.MustCreateMockBinaryMessage(v),
				want:    v,
			},
			{
				name:  "From event/" + test.NameOf(v),
				event: v,
				want:  v,
			},
		}

		for _, tt := range testCases {
			t.Run("With structured Transformer "+tt.name, func(t *testing.T) {
				testToEventWithTransformer(t, tt, test.NewMockTransformerFactory(false, false))
			})
			t.Run("With binary Transformer "+tt.name, func(t *testing.T) {
				testToEventWithTransformer(t, tt, test.NewMockTransformerFactory(true, false))
			})
			t.Run("With event Transformer "+tt.name, func(t *testing.T) {
				testToEventWithTransformer(t, tt, test.NewMockTransformerFactory(true, true))
			})
			t.Run("With mixed Transformers "+tt.name, func(t *testing.T) {
				var inputMessage binding.Message
				if tt.message != nil {
					inputMessage = tt.message
				} else {
					e := tt.event.Clone()
					inputMessage = binding.ToMessage(&e)
				}

				transformerBinary := test.NewMockTransformerFactory(true, false)
				transformerEvent := test.NewMockTransformerFactory(true, true)

				got, err := binding.ToEvent(context.Background(), inputMessage, transformerBinary, transformerEvent)
				require.NoError(t, err)
				test.AssertEventEquals(t, test.ExToStr(t, tt.want), test.ExToStr(t, *got))

				test.AssertTransformerInvokedOneTime(t, transformerBinary)
				require.Equal(t, 1, transformerBinary.InvokedEventTransformer)
				test.AssertTransformerInvokedOneTime(t, transformerEvent)
			})
		}
	})
}

func testToEventWithTransformer(t *testing.T, tt toEventTestCase, transformer *test.MockTransformerFactory) {
	var inputMessage binding.Message
	if tt.message != nil {
		inputMessage = tt.message
	} else {
		e := tt.event.Clone()
		inputMessage = binding.ToMessage(&e)
	}

	got, err := binding.ToEvent(context.Background(), inputMessage, transformer)
	require.NoError(t, err)
	test.AssertEventEquals(t, test.ExToStr(t, tt.want), test.ExToStr(t, *got))

	test.AssertTransformerInvokedOneTime(t, transformer)
}
