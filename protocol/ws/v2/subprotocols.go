package v2

import (
	"fmt"

	"nhooyr.io/websocket"

	"github.com/cloudevents/sdk-go/v2/binding/format"
)

const JsonSubprotocol = "cloudevents.json"

var SupportedSubprotocols = []string{JsonSubprotocol}

func resolveFormat(subprotocol string) (format.Format, websocket.MessageType, error) {
	switch subprotocol {
	case "cloudevents.json":
		return format.JSON, websocket.MessageText, nil
	default:
		return nil, websocket.MessageText, fmt.Errorf("subprotocol not supported: %s", subprotocol)
	}
}
