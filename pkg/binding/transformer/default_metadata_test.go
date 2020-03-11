package transformer

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/event"

	. "github.com/cloudevents/sdk-go/pkg/binding/test"
)

func TestAddUUID(t *testing.T) {
	eventWithoutId := test.MinEvent()
	eventCtx := eventWithoutId.Context.AsV1()
	eventCtx.ID = ""
	eventWithoutId.Context = eventCtx

	eventWithId := test.MinEvent()

	assertUUID := func(t *testing.T, ev event.Event) {
		require.NotZero(t, ev.Context.GetID())
		_, err := uuid.Parse(ev.Context.GetID())
		require.NoError(t, err)
	}

	RunTransformerTests(t, context.Background(), []TransformerTestArgs{
		{
			Name:         "No change to id to Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(eventWithId.Clone()),
			WantEvent:    eventWithId.Clone(),
			Transformers: []binding.TransformerFactory{AddUUID},
		},
		{
			Name:         "No change to id to Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(eventWithId.Clone()),
			WantEvent:    eventWithId.Clone(),
			Transformers: []binding.TransformerFactory{AddUUID},
		},
		{
			Name:         "No change to id to Event message",
			InputEvent:   eventWithId,
			WantEvent:    eventWithId,
			Transformers: []binding.TransformerFactory{AddUUID},
		},
		{
			Name:         "Add UUID to Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(eventWithoutId.Clone()),
			AssertFunc:   assertUUID,
			Transformers: []binding.TransformerFactory{AddUUID},
		},
		{
			Name:         "Add UUID to Event message",
			InputEvent:   eventWithoutId,
			AssertFunc:   assertUUID,
			Transformers: []binding.TransformerFactory{AddUUID},
		},
	})
}

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

	RunTransformerTests(t, context.Background(), []TransformerTestArgs{
		{
			Name:         "No change to time to Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(eventWithTime.Clone()),
			WantEvent:    eventWithTime.Clone(),
			Transformers: []binding.TransformerFactory{AddTimeNow},
		},
		{
			Name:         "No change to time to Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(eventWithTime.Clone()),
			WantEvent:    eventWithTime.Clone(),
			Transformers: []binding.TransformerFactory{AddTimeNow},
		},
		{
			Name:         "No change to time to Event message",
			InputEvent:   eventWithTime,
			WantEvent:    eventWithTime,
			Transformers: []binding.TransformerFactory{AddTimeNow},
		},
		{
			Name:         "Add time.Now() to Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(eventWithoutTime.Clone()),
			AssertFunc:   assertTimeNow,
			Transformers: []binding.TransformerFactory{AddTimeNow},
		},
		{
			Name:         "Add time.Now() to Event message",
			InputEvent:   eventWithoutTime,
			AssertFunc:   assertTimeNow,
			Transformers: []binding.TransformerFactory{AddTimeNow},
		},
	})
}
