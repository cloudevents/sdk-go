package mqttv3_paho

import (
	"context"
	"testing"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/stretchr/testify/require"
)

func TestProtocolMessage(t *testing.T) {
	ev := ce.NewEvent()
	ev.SetID("123")
	ev.SetSource("unit-test")
	ev.SetType("unit-test-type")
	ev.SetExtension("key", "value")

	err := ev.SetData(ce.ApplicationJSON, map[string]string{
		"hello": "world",
	})
	require.NoError(t, err)

	t.Run("marshal event", func(t *testing.T) {
		msg := (binding.EventMessage)(ev)

		b, err := WritePubMessage(context.Background(), &msg)
		require.NoError(t, err)
		require.NotEmpty(t, b)

		var result ce.Event
		err = result.UnmarshalJSON(b)
		require.NoError(t, err)

		require.Equal(t, ev, result)
	})

	t.Run("unmarshal event", func(t *testing.T) {
		payload, err := ev.MarshalJSON()
		require.NoError(t, err)

		msg := NewMessage(payload)

		result, err := binding.ToEvent(context.Background(), msg)
		require.NoError(t, err)

		require.Equal(t, &ev, result)
	})

	t.Run("unmarshal event gives valid encoding", func(t *testing.T) {
		payload, err := ev.MarshalJSON()
		require.NoError(t, err)

		msg := NewMessage(payload)
		require.Equal(t, binding.EncodingStructured, msg.ReadEncoding())
	})

	t.Run("unmarshal event gives unknown encoding for non ce JSON", func(t *testing.T) {
		msg := NewMessage([]byte(`{"hello": "world"}`))
		require.Equal(t, binding.EncodingUnknown, msg.ReadEncoding())
	})

	t.Run("unmarshal event gives unknown encoding for malformed payloads", func(t *testing.T) {
		msg := NewMessage([]byte(`{"`))
		require.Equal(t, binding.EncodingUnknown, msg.ReadEncoding())
	})
}
