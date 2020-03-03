package transcoder

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/binding/spec"
	"github.com/cloudevents/sdk-go/pkg/binding/test"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
)

func TestAddOrUpdateAttribute(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	e.Context.AsV01().EventTime = nil

	attributeKind := spec.Time
	attributeInitialValue := types.Timestamp{Time: time.Now().UTC()}
	attributeUpdatedValue := types.Timestamp{Time: attributeInitialValue.Add(1 * time.Hour)}

	eventWithInitialValue := test.CopyEventContext(e)
	eventWithInitialValue.SetTime(attributeInitialValue.Time)

	eventWithUpdatedValue := test.CopyEventContext(e)
	eventWithUpdatedValue.SetTime(attributeUpdatedValue.Time)

	transformers := AddOrUpdateAttribute(attributeKind, attributeInitialValue.Time, func(i2 interface{}) (i interface{}, err error) {
		require.NotNil(t, i2)
		t, err := types.ToTime(i2)
		if err != nil {
			return nil, err
		}

		return t.Add(1 * time.Hour), nil
	})

	test.RunTranscoderTests(t, context.Background(), []test.TranscoderTestArgs{
		{
			Name:         "Add time to Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(e),
			WantEvent:    eventWithInitialValue,
			Transformers: transformers,
		},
		{
			Name:         "Add time to Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(e),
			WantEvent:    eventWithInitialValue,
			Transformers: transformers,
		},
		{
			Name:         "Add time to Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(e)),
			WantEvent:    eventWithInitialValue,
			Transformers: transformers,
		},
		{
			Name:         "Update time in Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(eventWithInitialValue),
			WantEvent:    eventWithUpdatedValue,
			Transformers: transformers,
		},
		{
			Name:         "Update time in Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(eventWithInitialValue),
			WantEvent:    eventWithUpdatedValue,
			Transformers: transformers,
		},
		{
			Name:         "Update time in Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(eventWithInitialValue)),
			WantEvent:    eventWithUpdatedValue,
			Transformers: transformers,
		},
	})
}

// Test a common flow: If the metadata is not existing, initialize with a value. Otherwise, update it
func TestAddOrUpdateExtension(t *testing.T) {
	e := test.MinEvent()
	e.Context = e.Context.AsV1()

	extName := "exnum"
	extInitialValue := "1"
	exUpdatedValue := "2"

	eventWithInitialValue := test.CopyEventContext(e)
	require.NoError(t, eventWithInitialValue.Context.SetExtension(extName, extInitialValue))

	eventWithUpdatedValue := test.CopyEventContext(e)
	require.NoError(t, eventWithUpdatedValue.Context.SetExtension(extName, exUpdatedValue))

	transformers := AddOrUpdateExtension(extName, extInitialValue, func(i2 interface{}) (i interface{}, err error) {
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

	test.RunTranscoderTests(t, context.Background(), []test.TranscoderTestArgs{
		{
			Name:         "Add exnum to Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(e),
			WantEvent:    eventWithInitialValue,
			Transformers: transformers,
		},
		{
			Name:         "Add exnum to Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(e),
			WantEvent:    eventWithInitialValue,
			Transformers: transformers,
		},
		{
			Name:         "Add exnum to Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(e)),
			WantEvent:    eventWithInitialValue,
			Transformers: transformers,
		},
		{
			Name:         "Update exnum in Mock Structured message",
			InputMessage: test.NewMockStructuredMessage(eventWithInitialValue),
			WantEvent:    eventWithUpdatedValue,
			Transformers: transformers,
		},
		{
			Name:         "Update exnum in Mock Binary message",
			InputMessage: test.NewMockBinaryMessage(eventWithInitialValue),
			WantEvent:    eventWithUpdatedValue,
			Transformers: transformers,
		},
		{
			Name:         "Update exnum in Event message",
			InputMessage: binding.EventMessage(test.CopyEventContext(eventWithInitialValue)),
			WantEvent:    eventWithUpdatedValue,
			Transformers: transformers,
		},
	})
}
