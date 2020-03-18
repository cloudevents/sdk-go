package transformer

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/binding/test"
	"github.com/cloudevents/sdk-go/v2/types"

	. "github.com/cloudevents/sdk-go/v2/binding/test"
)

func TestSetAttribute(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	attributeKind := spec.Time
	attributeInitialValue := types.Timestamp{Time: time.Now().UTC()}
	attributeUpdatedValue := types.Timestamp{Time: attributeInitialValue.Add(1 * time.Hour)}

	eventWithInitialValue := e.Clone()
	eventWithInitialValue.SetTime(attributeInitialValue.Time)

	eventWithUpdatedValue := e.Clone()
	eventWithUpdatedValue.SetTime(attributeUpdatedValue.Time)

	transformers := SetAttribute(attributeKind, attributeInitialValue.Time, func(i2 interface{}) (i interface{}, err error) {
		require.NotNil(t, i2)
		t, err := types.ToTime(i2)
		if err != nil {
			return nil, err
		}

		return t.Add(1 * time.Hour), nil
	})

	RunTransformerTests(t, context.Background(), []TransformerTestArgs{
		{
			Name:         "Add time to Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(e),
			WantEvent:    eventWithInitialValue,
			Transformers: transformers,
		},
		{
			Name:         "Add time to Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(e),
			WantEvent:    eventWithInitialValue,
			Transformers: transformers,
		},
		{
			Name:         "Add time to Event message",
			InputEvent:   e,
			WantEvent:    eventWithInitialValue,
			Transformers: transformers,
		},
		{
			Name:         "Update time in Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(eventWithInitialValue),
			WantEvent:    eventWithUpdatedValue,
			Transformers: transformers,
		},
		{
			Name:         "Update time in Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(eventWithInitialValue),
			WantEvent:    eventWithUpdatedValue,
			Transformers: transformers,
		},
		{
			Name:         "Update time in Event message",
			InputEvent:   eventWithInitialValue,
			WantEvent:    eventWithUpdatedValue,
			Transformers: transformers,
		},
	})
}

// Test a common flow: If the metadata is not existing, initialize with a value. Otherwise, update it
func TestSetExtension(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	extName := "exnum"
	extInitialValue := "1"
	exUpdatedValue := "2"

	eventWithInitialValue := e.Clone()
	require.NoError(t, eventWithInitialValue.Context.SetExtension(extName, extInitialValue))

	eventWithUpdatedValue := e.Clone()
	require.NoError(t, eventWithUpdatedValue.Context.SetExtension(extName, exUpdatedValue))

	transformers := SetExtension(extName, extInitialValue, func(i2 interface{}) (i interface{}, err error) {
		require.NotNil(t, i2)
		str, err := types.Format(i2)
		if err != nil {
			return nil, err
		}

		n, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		n++
		return strconv.Itoa(n), nil
	})

	RunTransformerTests(t, context.Background(), []TransformerTestArgs{
		{
			Name:         "Add exnum to Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(e),
			WantEvent:    eventWithInitialValue,
			Transformers: transformers,
		},
		{
			Name:         "Add exnum to Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(e),
			WantEvent:    eventWithInitialValue,
			Transformers: transformers,
		},
		{
			Name:         "Add exnum to Event message",
			InputEvent:   e,
			WantEvent:    eventWithInitialValue,
			Transformers: transformers,
		},
		{
			Name:         "Update exnum in Mock Structured message",
			InputMessage: test.MustCreateMockStructuredMessage(eventWithInitialValue),
			WantEvent:    eventWithUpdatedValue,
			Transformers: transformers,
		},
		{
			Name:         "Update exnum in Mock Binary message",
			InputMessage: test.MustCreateMockBinaryMessage(eventWithInitialValue),
			WantEvent:    eventWithUpdatedValue,
			Transformers: transformers,
		},
		{
			Name:         "Update exnum in Event message",
			InputEvent:   eventWithInitialValue,
			WantEvent:    eventWithUpdatedValue,
			Transformers: transformers,
		},
	})
}
