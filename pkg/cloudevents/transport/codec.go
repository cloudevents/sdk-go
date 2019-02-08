package transport

import "github.com/cloudevents/sdk-go/pkg/cloudevents/canonical"

// Codec is the interface for transport codecs to convert between transport
// specific payloads and the Message interface.
type Codec interface {
	Encode(canonical.Event) (Message, error)
	Decode(Message) (canonical.Event, error)
}
