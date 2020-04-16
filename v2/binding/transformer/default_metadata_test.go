package transformer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/event"
)

func TestAddTimeNow(t *testing.T) {
	eventWithoutTime := test.MinEvent()
	eventCtx := eventWithoutTime.Context.AsV1()
	eventCtx.Time = nil
	eventWithoutTime.Context = eventCtx

	eventWithTime := test.MinEvent()
	eventWithTime.SetTime(time.Now().Add(2 * time.Hour).UTC())

	assertTimeNow := func(t *testing.T, ev event.Event) {
		require.False(t, ev.Context.GetTime().IsZero())
	}

	test.RunTransformerTests(t, context.Background(), []test.TransformerTestArgs{
		{
			Name:         "No change to time to Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(eventWithTime.Clone()),
			WantEvent:    eventWithTime.Clone(),
			Transformers: binding.Transformers{AddTimeNow},
		},
		{
			Name:         "No change to time to Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(eventWithTime.Clone()),
			WantEvent:    eventWithTime.Clone(),
			Transformers: binding.Transformers{AddTimeNow},
		},
		{
			Name:         "No change to time to Event message",
			InputEvent:   eventWithTime,
			WantEvent:    eventWithTime,
			Transformers: binding.Transformers{AddTimeNow},
		},
		{
			Name:         "Add time.Now() to Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(eventWithoutTime.Clone()),
			AssertFunc:   assertTimeNow,
			Transformers: binding.Transformers{AddTimeNow},
		},
		{
			Name:         "Add time.Now() to Event message",
			InputEvent:   eventWithoutTime,
			AssertFunc:   assertTimeNow,
			Transformers: binding.Transformers{AddTimeNow},
		},
	})
}
