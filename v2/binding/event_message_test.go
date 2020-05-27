package binding_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	bindingtest "github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/test"
)

type mockFormat struct {
	t             *testing.T
	expectedEvent event.Event
}

func (m *mockFormat) MediaType() string {
	return "application/cool-mediatype"
}

func (m *mockFormat) Marshal(have *event.Event) ([]byte, error) {
	test.AssertEventEquals(m.t, m.expectedEvent, *have)
	return []byte{}, nil
}

func (m *mockFormat) Unmarshal([]byte, *event.Event) error {
	m.t.Fatal("This should never be invoked")
	return nil
}

func TestStructuredModeCustomFormat(t *testing.T) {
	e := test.FullEvent()
	format := mockFormat{t: t, expectedEvent: e}

	ctx := binding.UseFormatForEvent(context.TODO(), &format)
	enc := bindingtest.MockStructuredMessage{}
	message := binding.EventMessage(e)

	err := message.ReadStructured(ctx, &enc)
	require.NoError(t, err)
	require.Equal(t, &format, enc.Format)
}

func TestEventMessage_ReadStructured(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, inputEvent event.Event) {
		eventMessage := binding.ToMessage(&inputEvent)
		outMessage := bindingtest.MockStructuredMessage{}

		require.NoError(t, eventMessage.ReadStructured(context.TODO(), &outMessage))

		outputEvent, err := binding.ToEvent(context.TODO(), &outMessage)
		require.NoError(t, err)
		require.NotNil(t, outputEvent)

		test.AssertEventEquals(t, test.ConvertEventExtensionsToString(t, inputEvent), test.ConvertEventExtensionsToString(t, *outputEvent))
	})
}

func TestEventMessage_ReadBinary(t *testing.T) {
	test.EachEvent(t, test.Events(), func(t *testing.T, inputEvent event.Event) {
		eventMessage := binding.ToMessage(&inputEvent)
		outMessage := bindingtest.MockBinaryMessage{}

		require.NoError(t, outMessage.Start(context.TODO()))

		require.NoError(t, eventMessage.ReadBinary(context.TODO(), &outMessage))

		outputEvent, err := binding.ToEvent(context.TODO(), &outMessage)
		require.NoError(t, err)
		require.NotNil(t, outputEvent)

		test.AssertEventEquals(t, inputEvent, *outputEvent)
	})
}
