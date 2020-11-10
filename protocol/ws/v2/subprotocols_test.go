package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nhooyr.io/websocket"

	"github.com/cloudevents/sdk-go/v2/binding/format"
)

func Test_resolveFormat(t *testing.T) {
	tests := []struct {
		name            string
		subprotocol     string
		wantFormat      format.Format
		wantMessageType websocket.MessageType
	}{{
		name:            "JSON subprotocol",
		subprotocol:     JsonSubprotocol,
		wantFormat:      format.JSON,
		wantMessageType: websocket.MessageText,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt, messageType, err := resolveFormat(tt.subprotocol)
			require.NoError(t, err)
			require.Equal(t, tt.wantFormat, fmt)
			require.Equal(t, tt.wantMessageType, messageType)
		})
	}
}

func Test_resolveFormat_error(t *testing.T) {
	_, _, err := resolveFormat("lalala")
	require.Error(t, err, "subprotocol not supported: lalala")
}
