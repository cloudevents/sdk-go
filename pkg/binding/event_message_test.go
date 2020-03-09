package binding_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/event"
)

type mockFormat struct {
	t             *testing.T
	expectedEvent event.Event
}

func (m *mockFormat) MediaType() string {
	return "application/cool-mediatype"
}

func (m *mockFormat) Marshal(have event.Event) ([]byte, error) {
	test.AssertEventEquals(m.t, m.expectedEvent, have)
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
	enc := test.MockStructuredMessage{}
	message := binding.EventMessage(e)

	err := message.ReadStructured(ctx, &enc)
	require.NoError(t, err)
	require.Equal(t, &format, enc.Format)
}
